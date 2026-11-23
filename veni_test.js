"use strict";

const test = require("node:test");
const assert = require("node:assert/strict");
const { Veni } = require("./veni.js");

class FakeNode {
  constructor(tagName) {
    this.tagName = tagName.toUpperCase();
  }
}

class FakeRoot {
  constructor(tags) {
    this.nodeType = 9;
    this.nodes = tags.map((tag) => new FakeNode(tag));
  }

  querySelectorAll() {
    return this.nodes;
  }
}

class FakeTemplate {
  constructor(attrs, childTags) {
    this.attrs = attrs || {};
    this.id = this.attrs.id || "";
    this.content = {
      querySelectorAll: () => (childTags || []).map((tag) => new FakeNode(tag)),
      querySelector: () => null,
    };
  }

  getAttribute(name) {
    return Object.prototype.hasOwnProperty.call(this.attrs, name)
      ? this.attrs[name]
      : null;
  }
}

function withBrowser(fn) {
  const previous = {
    customElements: global.customElements,
    HTMLElement: global.HTMLElement,
  };
  const defined = new Map();
  global.customElements = {
    get: (name) => defined.get(name) || null,
    define: (name, ctor) => {
      if (defined.has(name)) throw new Error("already defined: " + name);
      defined.set(name, ctor);
    },
  };
  global.HTMLElement = class HTMLElement {};
  try {
    fn(defined);
  } finally {
    for (const key of Object.keys(previous)) {
      if (previous[key] === undefined) delete global[key];
      else global[key] = previous[key];
    }
  }
}

test("discover returns custom tags only, deduplicated", () => {
  withBrowser(() => {
    const veni = new Veni();
    const root = new FakeRoot(["x-item-card", "DIV", "x-item-card", "x-feed"]);
    assert.deepEqual(veni.discover(root), ["x-item-card", "x-feed"]);
  });
});

test("init registers known tags and marks unknown ones pending", () => {
  withBrowser((defined) => {
    class ItemCard extends global.HTMLElement {}
    const veni = new Veni({ registry: { "x-item-card": ItemCard } });
    veni.init(new FakeRoot(["x-item-card", "x-unknown", "SECTION"]));

    assert.equal(defined.has("x-item-card"), true);
    assert.equal(veni.isDefined("x-item-card"), true);
    assert.deepEqual(veni.pending(), ["x-unknown"]);
    assert.equal(veni.isDefined("x-unknown"), false);
  });
});

test("autodefine stubs pending tags and clears the pending list", () => {
  withBrowser((defined) => {
    const veni = new Veni();
    veni.init(new FakeRoot(["x-unknown"]));
    assert.deepEqual(veni.autodefine(), ["x-unknown"]);
    assert.deepEqual(veni.pending(), []);
    assert.equal(defined.has("x-unknown"), true);
    assert.equal(veni.isDefined("x-unknown"), true);
  });
});

test("define validates names and defineAll registers every entry", () => {
  withBrowser(() => {
    const veni = new Veni();
    class Card extends global.HTMLElement {}
    class Feed extends global.HTMLElement {}

    assert.equal(veni.define("NoHyphen", Card), null);
    assert.equal(veni.define("x-", Card), null);
    assert.equal(veni.define("X-Item-Card", Card), "x-item-card");
    assert.deepEqual(veni.defineAll({ "x-feed": Feed }), ["x-feed"]);
    assert.deepEqual(veni.list(), ["x-feed", "x-item-card"]);
  });
});

test("fromTemplate derives a name from data-component or a t- prefixed id", () => {
  withBrowser(() => {
    const veni = new Veni();
    class Card extends global.HTMLElement {}

    const fromAttr = veni.fromTemplate(new FakeTemplate({ "data-component": "x-card" }), Card);
    const fromId = veni.fromTemplate(new FakeTemplate({ id: "t-item-card" }), Card);

    assert.equal(fromAttr, "x-card");
    assert.equal(fromId, "item-card");
    assert.equal(veni.isDefined("x-card"), true);
    assert.equal(veni.isDefined("item-card"), true);
  });
});

test("on wraps connectedCallback without replacing it", () => {
  withBrowser((defined) => {
    const veni = new Veni();
    const calls = [];
    class Widget extends global.HTMLElement {
      connectedCallback() {
        calls.push("original");
      }
    }
    veni.define("x-widget", Widget);
    veni.on("x-widget", "created", () => calls.push("hook"));

    new (defined.get("x-widget"))().connectedCallback();
    assert.deepEqual(calls, ["hook", "original"]);
  });
});
