import { Injectable, signal } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Flag, EvaluationContext, EvaluationResult } from '../models/flag.model';
import { finalize } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class FlagService {
  private apiUrl = environment.apiUrl;
  private headers = new HttpHeaders({
    'X-API-KEY': environment.apiKey,
    'Content-Type': 'application/json'
  });

  flags = signal<Flag[]>([]);
  loading = signal<boolean>(false);
  error = signal<string | null>(null);

  constructor(private http: HttpClient) {}

  loadFlags() {
    console.log('Loading flags from:', `${this.apiUrl}/flags`);
    this.loading.set(true);
    this.error.set(null);
    
    this.http.get<Flag[]>(`${this.apiUrl}/flags`, { headers: this.headers })
      .pipe(
        finalize(() => this.loading.set(false))
      )
      .subscribe({
        next: (flags) => {
          console.log('Successfully loaded flags:', flags.length);
          this.flags.set(flags);
        },
        error: (err) => {
          console.error('API Error:', err);
          this.error.set(`Failed to connect to API (${err.status}: ${err.statusText || 'Unknown Error'}). Ensure API is running at ${this.apiUrl}`);
        }
      });
  }

  getFlag(id: string) {
    return this.http.get<Flag>(`${this.apiUrl}/flags/${id}`, { headers: this.headers });
  }

  createFlag(flag: Flag) {
    return this.http.post<Flag>(`${this.apiUrl}/flags`, flag, { headers: this.headers });
  }

  updateFlag(id: string, flag: Partial<Flag>) {
    return this.http.patch<Flag>(`${this.apiUrl}/flags/${id}`, flag, { headers: this.headers });
  }

  deleteFlag(id: string) {
    return this.http.delete(`${this.apiUrl}/flags/${id}`, { headers: this.headers });
  }

  evaluate(key: string, context: EvaluationContext) {
    return this.http.post<EvaluationResult>(`${this.apiUrl}/flags/${key}/evaluate`, context, { headers: this.headers });
  }
}
