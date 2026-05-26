import { Component, OnInit } from '@angular/core';
import { CommonModule, DatePipe } from '@angular/common';
import { RouterLink } from '@angular/router';
import { FlagService } from '../../services/flag.service';

@Component({
  selector: 'app-flag-list',
  standalone: true,
  imports: [CommonModule, RouterLink, DatePipe],
  template: `
    <div class="header">
      <div>
        <h1>Feature Flags</h1>
        <p class="text-muted">Manage your deployment controls and rollout rules.</p>
      </div>
      <button class="btn btn-primary" routerLink="/flags/new">Create Flag</button>
    </div>

    <div class="card">
      @if (flagService.loading()) {
        <div class="loading">
          <div class="spinner"></div>
          <span>Loading flags...</span>
        </div>
      } @else if (flagService.error()) {
        <div class="error-container">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          <p>{{ flagService.error() }}</p>
          <button class="btn btn-sm" (click)="flagService.loadFlags()">Try Again</button>
        </div>
      } @else {
        <table class="table">
          <thead>
            <tr>
              <th>Status</th>
              <th>Name & Key</th>
              <th>Strategy</th>
              <th>Rules</th>
              <th>Last Updated</th>
              <th>Updated By</th>
              <th class="text-right">Actions</th>
            </tr>
          </thead>
          <tbody>
            @for (flag of flagService.flags(); track flag.key) {
              <tr>
                <td style="width: 100px;">
                  <span class="pill" [ngClass]="flag.enabled ? 'pill-success' : 'pill-muted'">
                    {{ flag.enabled ? 'Active' : 'Disabled' }}
                  </span>
                </td>
                <td>
                  <div class="font-medium">{{ flag.name }}</div>
                  <div class="text-xs text-muted">{{ flag.key }}</div>
                </td>
                <td>
                  <span class="strategy-badge">{{ flag.ruleMatchStrategy }}</span>
                </td>
                <td>
                  <span class="rules-count">{{ flag.rules ? flag.rules.length : 0 }}</span>
                </td>
                <td class="text-sm">
                  {{ (flag.updatedAt || flag.createdAt) | date:'MMM d, HH:mm' }}
                </td>
                <td class="text-sm">
                  {{ flag.updatedBy || flag.createdBy || 'System' }}
                </td>
                <td class="text-right">
                  <a [routerLink]="['/flags', flag.id]" class="btn-action">Edit</a>
                </td>
              </tr>
            } @empty {
              <tr>
                <td colspan="7" class="empty-state">
                  <p>No feature flags found.</p>
                  <a routerLink="/flags/new" class="text-primary">Create your first flag</a>
                </td>
              </tr>
            }
          </tbody>
        </table>
      }
    </div>
  `,
  styles: [`
    .header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: var(--space-xl);
    }
    .header h1 { margin-bottom: 4px; }
    .btn {
      padding: 0.625rem 1.25rem;
      border-radius: 6px;
      font-weight: 500;
      text-decoration: none;
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
    .btn-primary {
      background: var(--primary);
      color: white;
    }
    .btn-primary:hover { background: var(--primary-hover); }
    .btn-sm { padding: 0.375rem 0.75rem; font-size: 0.875rem; background: white; border: 1px solid var(--border); margin-top: 12px; }
    
    .table {
      width: 100%;
      border-collapse: collapse;
    }
    th {
      text-align: left;
      font-size: 0.75rem;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: var(--text-muted);
      padding: var(--space-md);
      border-bottom: 1px solid var(--border);
      font-weight: 600;
    }
    td {
      padding: var(--space-md);
      border-bottom: 1px solid var(--border);
      vertical-align: middle;
    }
    .text-right { text-align: right; }
    .font-medium { font-weight: 500; color: var(--text-main); }
    .text-xs { font-size: 0.75rem; }
    .text-sm { font-size: 0.875rem; color: var(--text-muted); }
    .strategy-badge {
      background: #f1f5f9;
      color: #475569;
      padding: 2px 8px;
      border-radius: 4px;
      font-size: 0.75rem;
      font-family: ui-monospace, monospace;
      font-weight: 600;
      text-transform: uppercase;
    }
    .rules-count {
      background: #eff6ff;
      color: var(--primary);
      width: 24px;
      height: 24px;
      display: inline-flex;
      align-items: center;
      justify-content: center;
      border-radius: 50%;
      font-size: 0.75rem;
      font-weight: 600;
    }
    .btn-action {
      color: var(--primary);
      text-decoration: none;
      font-size: 0.875rem;
      font-weight: 600;
      padding: 6px 12px;
      border-radius: 4px;
      transition: background 0.2s;
    }
    .btn-action:hover { background: #f0f7ff; }
    
    .loading {
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 12px;
      padding: var(--space-xl);
      color: var(--text-muted);
    }
    .spinner {
      width: 24px;
      height: 24px;
      border: 3px solid #e2e8f0;
      border-top-color: var(--primary);
      border-radius: 50%;
      animation: spin 0.8s linear infinite;
    }
    @keyframes spin { to { transform: rotate(360deg); } }

    .error-container {
      padding: var(--space-xl);
      text-align: center;
      color: var(--danger);
    }
    .error-container svg { width: 48px; height: 48px; margin-bottom: 12px; opacity: 0.5; }
    
    .empty-state {
      padding: 4rem var(--space-xl);
      text-align: center;
      color: var(--text-muted);
    }
    .empty-state p { margin-bottom: 8px; font-weight: 500; }
    .text-primary { color: var(--primary); font-weight: 600; text-decoration: none; }
  `]
})
export class FlagListComponent implements OnInit {
  constructor(public flagService: FlagService) {}

  ngOnInit() {
    this.flagService.loadFlags();
  }
}
