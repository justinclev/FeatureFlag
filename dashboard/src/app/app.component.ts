import { Component } from '@angular/core';
import { RouterOutlet, RouterLink, RouterLinkActive } from '@angular/router';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive],
  template: `
    <div class="layout">
      <aside class="sidebar">
        <div class="logo">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
          </svg>
          <span>FeatureFlag</span>
          <span class="version-tag">v2.0</span>
        </div>
        <nav>
          <a routerLink="/flags" routerLinkActive="active" [routerLinkActiveOptions]="{exact: true}">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="nav-icon">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/><line x1="9" y1="3" x2="9" y2="21"/>
            </svg>
            Flags
          </a>
        </nav>
        <div class="sidebar-footer">
          <div class="status-indicator">
            <div class="dot"></div>
            System Online
          </div>
          <div class="build-info">
            Build: 2026-05-25 21:58
          </div>
        </div>
      </aside>
      <main class="content">
        <router-outlet></router-outlet>
      </main>
    </div>
  `,
  styles: [`
    .layout {
      display: grid;
      grid-template-columns: 260px 1fr;
      min-height: 100vh;
    }
    .sidebar {
      background: #0f172a;
      color: #f8fafc;
      padding: var(--space-xl) var(--space-lg);
      display: flex;
      flex-direction: column;
      gap: 2rem;
      position: sticky;
      top: 0;
      height: 100vh;
    }
    .logo {
      display: flex;
      align-items: center;
      gap: 12px;
      font-size: 1.25rem;
      font-weight: 700;
      color: white;
    }
    .logo svg { width: 28px; height: 28px; color: #3b82f6; }
    .version-tag {
      font-size: 0.65rem;
      background: #1e293b;
      color: #94a3b8;
      padding: 2px 6px;
      border-radius: 4px;
      margin-left: auto;
    }
    nav {
      display: flex;
      flex-direction: column;
      gap: 4px;
    }
    nav a {
      text-decoration: none;
      color: #94a3b8;
      padding: 0.75rem 1rem;
      border-radius: 8px;
      font-weight: 500;
      display: flex;
      align-items: center;
      gap: 12px;
      transition: all 0.2s;
    }
    .nav-icon { width: 18px; height: 18px; }
    nav a:hover {
      background: rgba(255, 255, 255, 0.05);
      color: white;
    }
    nav a.active {
      background: #3b82f6;
      color: white;
    }
    .sidebar-footer {
      margin-top: auto;
      padding-top: var(--space-lg);
      border-top: 1px solid #1e293b;
    }
    .status-indicator {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 0.75rem;
      color: #94a3b8;
      margin-bottom: 4px;
    }
    .build-info {
      font-size: 0.6rem;
      color: #475569;
      font-family: monospace;
    }
    .dot {
      width: 8px;
      height: 8px;
      background: #22c55e;
      border-radius: 50%;
      box-shadow: 0 0 8px rgba(34, 197, 94, 0.5);
    }
    .content {
      background: #f8fafc;
      padding: 2.5rem;
      overflow-y: auto;
    }
  `]
})
export class AppComponent {}
