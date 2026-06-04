import { Injectable, signal, computed } from '@angular/core';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { environment } from '../../environments/environment';
import { Flag, EvaluationContext, EvaluationResult } from '../domain/models/flag.domain';
import { finalize, tap } from 'rxjs';

@Injectable({ providedIn: 'root' })
export class FlagService {
  private apiUrl = environment.apiUrl;
  private headers = new HttpHeaders({
    'X-API-KEY': environment.apiKey,
    'Content-Type': 'application/json'
  });

  // --- State ---
  private flagsState = signal<Flag[]>([]);
  private loadingState = signal<boolean>(false);
  private errorState = signal<string | null>(null);

  // --- Selectors ---
  flags = computed(() => this.flagsState());
  loading = computed(() => this.loadingState());
  error = computed(() => this.errorState());
  activeCount = computed(() => this.flagsState().filter(f => f.enabled).length);

  constructor(private http: HttpClient) {}

  // --- Actions ---
  loadFlags() {
    this.loadingState.set(true);
    this.errorState.set(null);
    
    return this.http.get<Flag[]>(`${this.apiUrl}/flags`, { headers: this.headers })
      .pipe(
        finalize(() => this.loadingState.set(false)),
        tap({
          next: (flags) => this.flagsState.set(flags),
          error: (err) => this.errorState.set(`Connection failed: ${err.status}`)
        })
      );
  }

  getFlag(id: string) {
    return this.http.get<Flag>(`${this.apiUrl}/flags/${id}`, { headers: this.headers });
  }

  createFlag(flag: Flag) {
    return this.http.post<Flag>(`${this.apiUrl}/flags`, flag, { headers: this.headers }).pipe(
      tap(newFlag => this.flagsState.update(list => [...list, newFlag]))
    );
  }

  updateFlag(id: string, flag: Partial<Flag>) {
    return this.http.patch<Flag>(`${this.apiUrl}/flags/${id}`, flag, { headers: this.headers }).pipe(
      tap(updated => this.flagsState.update(list => list.map(f => f.id === id ? updated : f)))
    );
  }

  deleteFlag(id: string) {
    return this.http.delete(`${this.apiUrl}/flags/${id}`, { headers: this.headers }).pipe(
      tap(() => this.flagsState.update(list => list.filter(f => f.id !== id)))
    );
  }

  evaluate(key: string, context: EvaluationContext) {
    return this.http.post<EvaluationResult>(`${this.apiUrl}/flags/${key}/evaluate`, context, { headers: this.headers });
  }
}
