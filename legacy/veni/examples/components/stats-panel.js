/**
 * stats-panel.js
 * 
 * A statistics dashboard component for displaying metrics.
 * 
 * Attributes:
 *   - total-users (string): Total number of users
 *   - active-users (string): Currently active users
 *   - sessions-today (string): Number of sessions today
 *   - avg-session (string): Average session duration
 * 
 * Usage:
 *   <stats-panel total-users="1,247" active-users="89" sessions-today="342"></stats-panel>
 */

class StatsPanel extends HTMLElement {
  constructor() {
    super();
    // Attach Shadow DOM with open mode
    this.attachShadow({ mode: 'open' });
  }

  /**
   * Called when the element is inserted into the DOM.
   * Renders the component's content based on current attributes.
   */
  connectedCallback() {
    this.render();
    
    // Observe attribute changes to re-render dynamically
    this.observer = new MutationObserver(() => this.render());
    this.observer.observe(this, { 
      attributes: true, 
      attributeFilter: ['total-users', 'active-users', 'sessions-today', 'avg-session'] 
    });
  }

  /**
   * Called when the element is removed from the DOM.
   * Cleans up observers.
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
    const totalUsers = this.getAttribute('total-users') || '0';
    const activeUsers = this.getAttribute('active-users') || '0';
    const sessionsToday = this.getAttribute('sessions-today') || '0';
    const avgSession = this.getAttribute('avg-session') || '0';

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }

        .panel {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          border-radius: 16px;
          padding: 2rem;
          box-shadow: 0 10px 40px rgba(102, 126, 234, 0.3);
          color: white;
        }

        .panel-title {
          font-size: 1.25rem;
          font-weight: 700;
          margin: 0 0 1.5rem 0;
          text-align: center;
          opacity: 0.9;
        }

        .stats-grid {
          display: grid;
          grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
          gap: 1.5rem;
        }

        .stat-item {
          background: rgba(255, 255, 255, 0.15);
          backdrop-filter: blur(10px);
          border-radius: 12px;
          padding: 1.25rem;
          text-align: center;
          transition: transform 0.2s ease, background 0.2s ease;
        }

        .stat-item:hover {
          transform: translateY(-4px);
          background: rgba(255, 255, 255, 0.25);
        }

        .stat-value {
          font-size: 2rem;
          font-weight: 800;
          margin: 0 0 0.25rem 0;
          line-height: 1.2;
        }

        .stat-label {
          font-size: 0.75rem;
          font-weight: 600;
          text-transform: uppercase;
          letter-spacing: 1px;
          opacity: 0.8;
          margin: 0;
        }

        /* Active users highlight */
        .stat-item.active {
          background: rgba(255, 255, 255, 0.25);
          border: 2px solid rgba(255, 255, 255, 0.3);
        }

        .stat-item.active .stat-value {
          color: #ffd700;
        }

        /* Pulse animation for active count */
        @keyframes pulse {
          0%, 100% { opacity: 1; }
          50% { opacity: 0.7; }
        }

        .stat-item.active .stat-value {
          animation: pulse 2s infinite;
        }

        /* Responsive adjustments */
        @media (max-width: 600px) {
          .panel {
            padding: 1.5rem;
          }

          .stats-grid {
            grid-template-columns: repeat(2, 1fr);
            gap: 1rem;
          }

          .stat-value {
            font-size: 1.5rem;
          }

          .stat-label {
            font-size: 0.65rem;
          }
        }

        @media (max-width: 350px) {
          .stats-grid {
            grid-template-columns: 1fr;
          }
        }
      </style>

      <div class="panel">
        <h3 class="panel-title">📊 Dashboard Statistics</h3>
        
        <div class="stats-grid">
          <div class="stat-item">
            <div class="stat-value">${this.escapeHtml(totalUsers)}</div>
            <p class="stat-label">Total Users</p>
          </div>

          <div class="stat-item active">
            <div class="stat-value">${this.escapeHtml(activeUsers)}</div>
            <p class="stat-label">● Active Now</p>
          </div>

          <div class="stat-item">
            <div class="stat-value">${this.escapeHtml(sessionsToday)}</div>
            <p class="stat-label">Sessions Today</p>
          </div>

          <div class="stat-item">
            <div class="stat-value">${this.escapeHtml(avgSession)}</div>
            <p class="stat-label">Avg Session</p>
          </div>
        </div>
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

// Register the custom element
customElements.define('stats-panel', StatsPanel);