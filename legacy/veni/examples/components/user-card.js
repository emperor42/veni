/**
 * user-card.js
 * 
 * A reusable user profile card component.
 * 
 * Attributes:
 *   - name (string): The user's display name (required)
 *   - email (string): The user's email address
 *   - role (string): The user's job title or role
 *   - avatar (string): URL to the user's avatar image
 * 
 * Slots:
 *   - bio (optional): Custom bio content to display below the email
 * 
 * Usage:
 *   <user-card name="Alice" email="alice@example.com" role="Developer"></user-card>
 */

class UserCard extends HTMLElement {
  constructor() {
    super();
    // Attach Shadow DOM with open mode to allow external access if needed
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
    this.observer.observe(this, { attributes: true, attributeFilter: ['name', 'email', 'role', 'avatar'] });
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
    const name = this.getAttribute('name') || 'Unknown User';
    const email = this.getAttribute('email') || '';
    const role = this.getAttribute('role') || '';
    const avatarUrl = this.getAttribute('avatar');

    // Generate initials for placeholder avatar
    const initials = this.getInitials(name);

    // Determine avatar HTML
    const avatarHtml = avatarUrl
      ? `<img src="$${this.escapeHtml(avatarUrl)}" alt="$${this.escapeHtml(name)}" class="avatar">`
      : `<div class="avatar-placeholder" aria-hidden="true">${initials}</div>`;

    // Determine role HTML
    const roleHtml = role ? `<span class="role-badge">${this.escapeHtml(role)}</span>` : '';

    // Determine bio slot content
    const slotContent = this.querySelector('[slot="bio"]');
    const bioHtml = slotContent 
      ? `<div class="bio-slot"><slot name="bio"></slot></div>`
      : '';

    this.shadowRoot.innerHTML = `
      <style>
        :host {
          display: block;
          font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
        }

        .card {
          display: flex;
          align-items: flex-start;
          gap: 1rem;
          padding: 1.25rem;
          background: #ffffff;
          border: 1px solid #e0e0e0;
          border-radius: 12px;
          box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
          transition: transform 0.2s ease, box-shadow 0.2s ease;
          max-width: 100%;
        }

        .card:hover {
          transform: translateY(-2px);
          box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
        }

        .avatar-container {
          flex-shrink: 0;
          position: relative;
        }

        .avatar, .avatar-placeholder {
          width: 64px;
          height: 64px;
          border-radius: 50%;
          object-fit: cover;
          border: 2px solid #f0f0f0;
        }

        .avatar-placeholder {
          background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
          color: white;
          display: flex;
          align-items: center;
          justify-content: center;
          font-weight: 700;
          font-size: 1.5rem;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .info {
          flex-grow: 1;
          min-width: 0; /* Prevents overflow */
        }

        .name {
          font-size: 1.125rem;
          font-weight: 700;
          color: #1a1a1a;
          margin: 0 0 0.25rem 0;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .email {
          font-size: 0.875rem;
          color: #666;
          margin: 0 0 0.5rem 0;
          white-space: nowrap;
          overflow: hidden;
          text-overflow: ellipsis;
        }

        .role-badge {
          display: inline-block;
          padding: 0.25rem 0.75rem;
          background: #f0f4ff;
          color: #4f46e5;
          font-size: 0.75rem;
          font-weight: 600;
          border-radius: 9999px;
          text-transform: uppercase;
          letter-spacing: 0.5px;
        }

        .bio-slot {
          margin-top: 0.75rem;
          padding-top: 0.75rem;
          border-top: 1px solid #f0f0f0;
          font-size: 0.875rem;
          color: #444;
          line-height: 1.5;
        }

        /* Responsive adjustments */
        @media (max-width: 480px) {
          .card {
            flex-direction: column;
            align-items: center;
            text-align: center;
          }
          
          .info {
            text-align: center;
          }
        }
      </style>

      <div class="card">
        <div class="avatar-container">
          ${avatarHtml}
        </div>
        
        <div class="info">
          <h3 class="name">${this.escapeHtml(name)}</h3>
          $${email ? `<p class="email">$${this.escapeHtml(email)}</p>` : ''}
          ${roleHtml}
          ${bioHtml}
        </div>
      </div>
    `;
  }

  /**
   * Generates initials from a name string.
   * Example: "Alice Johnson" -> "AJ"
   */
  getInitials(name) {
    return name
      .split(' ')
      .filter(part => part.length > 0)
      .map(part => part[0])
      .join('')
      .toUpperCase()
      .slice(0, 2);
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
customElements.define('user-card', UserCard);