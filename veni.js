"use strict";

/**
 * veni — web-component identification & management.
 *
 * Discovers candidate custom elements in the document and in <template>
 * fragments, registers them through customElements.define, and keeps a small
 * registry so generated pages can reuse components without duplicating code.
 *
 * Usage:
 *   veni.init();                       // discover + register on load (automatic)
 *   veni.define("x-item-card", ItemCard);   // register a named component
 *   var pending = veni.discover(root);      // list not-yet-defined tags
 *   veni.defineAll({ "x-item-card": ItemCard, "x-feed": FeedCard });
 *
 * Browser global: window.veni (instance) and window.Veni (class).
 * No external dependencies; plain web platform only.
 */

(function () {
  "use strict";

  // A valid custom-element name: lowercase, starts with a letter, contains
  // at least one hyphen, and ends alphanumeric. This is the exact constraint
  // customElements.define enforces, so anything we register must match.
  var TAG_RE = /^[a-z][a-z0-9]*-[a-z0-9-]*$/;

  function isCustomTag(name) {
    return TAG_RE.test(name);
  }

  function tagNameOf(node) {
    if (!node || !node.tagName) return "";
    var t = String(node.tagName).toLowerCase();
    return isCustomTag(t) ? t : "";
  }

  class Veni {
    constructor(opts) {
      opts = opts || {};
      this.registry = opts.registry || {}; // name -> constructor
      this.handlers = opts.handlers || {}; // name -> { created, connected, disconnected } callbacks
      this._defined = new Set();
      this._pending = new Set(); // discovered, not yet defined
      this._initiated = false;
    }

    /**
     * Walk the document (and template contents) and register every
     * custom-element tag that appears *and has a registered constructor*,
     * upgrading existing elements in place. Unknown tags are recorded as
     * pending (see pending()/autodefine()) so application code can still
     * provide the real constructor — init() never silently stubs a tag that
     * a later script may want to define properly.
     */
    init(root) {
      var docs = root && root.nodeType === 9 ? root : document;
      var roots = [docs];

      if (docs.querySelectorAll) {
        var templates = docs.querySelectorAll("template");
        for (var i = 0; i < templates.length; i++) {
          if (templates[i].content) roots.push(templates[i].content);
        }
      }

      var seen = {};
      for (var r = 0; r < roots.length; r++) {
        if (!roots[r].querySelectorAll) continue;
        var nodes = roots[r].querySelectorAll("*");
        for (var k = 0; k < nodes.length; k++) {
          var tn = tagNameOf(nodes[k]);
          if (!tn || seen[tn]) continue;
          seen[tn] = true;
          if (customElements.get(tn)) {
            this._defined.add(tn);
          } else if (this.registry[tn]) {
            this.registerExisting(tn);
          } else {
            this._pending.add(tn);
          }
        }
      }

      this._initiated = true;
      return this;
    }

    /** Tags discovered but not yet defined (application code can define them). */
    pending() {
      return Array.from(this._pending).sort();
    }

    /**
     * Define a lightweight stub for every pending tag. Useful for pages that
     * want "unknown tags upgrade to containers" behavior without providing
     * real constructors. Returns the list of names defined.
     */
    autodefine() {
      var out = [];
      var self = this;
      Array.from(this._pending).forEach(function (name) {
        if (customElements.get(name)) {
          self._pending.delete(name);
          self._defined.add(name);
          return;
        }
        try {
          customElements.define(name, self._stub(name));
          self._defined.add(name);
          self._pending.delete(name);
          out.push(name);
        } catch (e) { /* leave pending */ }
      });
      return out;
    }

    /**
     * Ensure a tag is defined using its registered constructor. Never stubs:
     * if there is no constructor, the tag stays pending for the app to define.
     */
    registerExisting(name) {
      if (!isCustomTag(name)) return null;
      if (customElements.get(name)) {
        this._defined.add(name);
        this._pending.delete(name);
        return name;
      }
      var ctor = this.registry[name];
      if (!ctor) return null;
      try {
        customElements.define(name, ctor);
        this._defined.add(name);
        this._pending.delete(name);
        this._wireHooks(name, ctor);
        return name;
      } catch (e) {
        if (customElements.get(name)) {
          this._defined.add(name);
          this._pending.delete(name);
          return name;
        }
        return null;
      }
    }

    _stub(name) {
      return class extends HTMLElement {
        connectedCallback() {
          if (!this.shadowRoot) {
            try {
              this.attachShadow({ mode: "open" });
            } catch (e) {
              /* shadow DOM unsupported — element still works */
              return;
            }
          }
          if (this.shadowRoot && !this.shadowRoot.firstChild) {
            this.shadowRoot.innerHTML = "<slot></slot>";
          }
        }
      };
    }

    /**
     * Register a named component with an explicit constructor. The fastest
     * path for application code: define("x-item-card", ItemCard).
     * Returns the name on success, null on failure.
     */
    define(name, ctor) {
      var t = String(name).toLowerCase();
      if (!isCustomTag(t)) return null;
      this.registry[t] = ctor;
      if (customElements.get(t)) {
        this._defined.add(t);
        return t;
      }
      try {
        customElements.define(t, ctor);
        this._defined.add(t);
        this._wireHooks(t, ctor);
        return t;
      } catch (e) {
        return customElements.get(t) ? t : null;
      }
    }

    register(name, ctor) {
      return this.define(name, ctor);
    }

    /** register many at once: { "x-item-card": Ctor, ... } -> array of names */
    defineAll(map) {
      var out = [];
      for (var k in map) {
        if (Object.prototype.hasOwnProperty.call(map, k)) {
          var n = this.define(k, map[k]);
          if (n) out.push(n);
        }
      }
      return out;
    }

    /**
     * attach a constructor to template-derived markup. The common failure mode
     * of old template components was naming them after the TEMPLATE/SLOT tags
     * themselves; this helper derives a compliant name from `id` or a
     * data-component attribute and registers it.
     */
    fromTemplate(templateEl, ctor) {
      var name = "";
      if (templateEl) {
        var base = templateEl.getAttribute("data-component") || templateEl.getAttribute("data-name");
        if (!base && templateEl.content) {
          // Common pattern: the template *wraps* a component, so the
          // data-component attribute sits on a child element.
          var first = templateEl.content.querySelector("[data-component], [data-name]");
          if (first) base = first.getAttribute("data-component") || first.getAttribute("data-name");
        }
        if (base) {
          base = String(base).replace(/\s+/g, "").toLowerCase();
        } else if (templateEl.id) {
          // Convention: <template id="t-item-card"> -> x-item-card.
          base = String(templateEl.id).replace(/\s+/g, "").toLowerCase();
          if (base.indexOf("t-") === 0) base = base.slice(2);
        }
      }
      if (!base || !isCustomTag(base)) {
        var fallback = "x-" + ((templateEl && templateEl.id) ? String(templateEl.id).replace(/\s+/g, "").toLowerCase() : "component");
        fallback = fallback.replace(/[^a-z0-9-]/g, "").replace(/-+/g, "-");
        if (!isCustomTag(fallback)) fallback = "x-component";
        base = fallback;
      }
      return this.define(base, ctor || this._stub(base));
    }

    /** Return all defined custom-element names. */
    list() {
      return Array.from(this._defined).sort();
    }

    /** Return tags present in a root that are not yet defined. */
    discover(root) {
      root = root || document;
      var out = [];
      if (!root.querySelectorAll) return out;
      var nodes = root.querySelectorAll("*");
      for (var i = 0; i < nodes.length; i++) {
        var t = tagNameOf(nodes[i]);
        if (t && !customElements.get(t) && out.indexOf(t) === -1) out.push(t);
      }
      return out;
    }

    /** true if a tag is defined (in the registry or via customElements). */
    isDefined(name) {
      return !!customElements.get(String(name).toLowerCase());
    }

    on(name, event, cb) {
      name = String(name).toLowerCase();
      this.handlers[name] = this.handlers[name] || {};
      this.handlers[name][event] = cb;
      var ctor = customElements.get(name);
      if (ctor) this._wireHooks(name, ctor);
      return this;
    }

    _wireHooks(name, ctor) {
      var h = this.handlers[name];
      if (!h) return;
      var proto = ctor.prototype;
      if (h.created && !proto.__veni_created) {
        var base = proto.connectedCallback || function () {};
        proto.__veni_created = true;
        proto.connectedCallback = function () {
          h.created && h.created(this);
          return base.call(this);
        };
      }
    }
  }

  var veni = new Veni();

  function boot() {
    try {
      veni.init(document);
    } catch (e) {
      if (typeof console !== "undefined") console.warn("veni.init:", e);
    }
  }

  if (typeof document !== "undefined") {
    if (document.readyState === "loading") {
      document.addEventListener("DOMContentLoaded", boot);
    } else {
      boot();
    }
  }

  // Export for module systems
  if (typeof module !== "undefined" && module.exports) {
    module.exports = { Veni: Veni, veni: veni };
  }

  if (typeof window !== "undefined") {
    window.Veni = Veni;
    window.veni = veni;
  }
})();