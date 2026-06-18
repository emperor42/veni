/**
 * hello-world.js
 * 
 * A simple greeting component that displays a welcome message.
 * 
 * Attributes:
 *   - name (string): The name to greet (defaults to "Guest")
 * 
 * Usage:
 *   <hello-world name="Alice"></hello-world>
 *   <hello-world></hello-world>
 */

class HelloWorld extends HTMLElement {
  constructor() {
    super();
    // Attach Shadow DOM with open mode
    // This encapsulates the component's styles and markup
    this.attachShadow({ mode: 'open' });
  }

  /**
   * Called when the element is inserted into the DOM.
   * Triggers the initial render.
   */
  connectedCallback() {
    this.render();
    
    // Observe attribute changes to update the greeting dynamically
    this.observer = new MutationObserver(() => this.render());
    this.observer.observe(this, { attributes: true, attributeFilter: ['name'] });
  }

  /**
   * Called when the element is removed from the DOM.
   * Cleans up the observer to prevent memory leaks.
   */
  disconnectedCallback() {
    if (this.observer) {
      this.observer.disconnect();
    }
  }

  /**
   * Renders the component's HTML and styles into the Shadow DOM.
   */
  render() {
    // Get the name attribute, default to "Guest" if not provided
    const name = this.getAttribute('name') || 'Guest';
    
    // Escape HTML to prevent XSS if the name comes from user input
    const safeName = this.escapeHtml(name);

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }

        .greeting-card {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          padding: 1.5rem 2rem;
          border-radius: 12px;
          box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
          text-align: center;
          transition: transform 0.3s ease, box-shadow 0.3s ease;
          position: relative;
          overflow: hidden;
        }

        .greeting-card:hover {
          transform: translateY(-5px);
          box-shadow: 0 8px 25px rgba(102, 126, 234, 0.6);
        }

        /* Subtle shine effect */
        .greeting-card::before {
          content: '';
          position: absolute;
          top: -50%;
          left: -50%;
          width: 200%;
          height: 200%;
          background: linear-gradient(
            to bottom right,
            rgba(255, 255, 255, 0) 0%,
            rgba(255, 255, 255, 0.1) 50%,
            rgba(255, 255, 255, 0) 100%
          );
          transform: rotate(45deg);
          pointer-events: none;
        }

        h2 {
          margin: 0 0 0.5rem 0;
          font-size: 1.75rem;
          font-weight: 700;
          letter-spacing: -0.5px;
        }

        p {
          margin: 0;
          font-size: 1rem;
          opacity: 0.9;
          font-weight: 300;
        }

        .emoji {
          font-size: 1.5rem;
          margin-right: 0.5rem;
          vertical-align: middle;
        }
      </style>

      <div class="greeting-card">
        <h2>
          <span class="emoji">👋</span>
          Hello, ${safeName}!
        </h2>
        <p>Welcome to the VENI ecosystem.</p>
      </div>
    `;
  }

  /**
   * Escapes HTML special characters to prevent XSS.
   */
  escapeHtml(text) {
    if (!text) return '';
    const map = {
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#039;'
    };
    return text.replace(/[&<>"']/g, m => map[m]);
  }
}

// Register the custom element with the browser
customElements.define('hello-world', HelloWorld);