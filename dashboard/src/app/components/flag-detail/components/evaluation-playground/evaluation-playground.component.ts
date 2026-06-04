import { Component, Input, signal, effect, model } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { FlagService } from '../../../../services/flag.service';
import { EvaluationResult, EvaluationContext } from '../../../../domain/models/flag.domain';

@Component({
  selector: 'app-evaluation-playground',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './evaluation-playground.component.html',
  styleUrl: './evaluation-playground.component.css'
})
export class EvaluationPlaygroundComponent {
  flagKey = model.required<string>();

  testContext = signal<EvaluationContext>({ 
    userId: 'user-123', 
    country: 'US', 
    state: '', 
    city: '', 
    zipCode: '', 
    attributes: {} 
  });
  
  testAttributes = signal<string>('{}');
  testResult = signal<EvaluationResult | null>(null);
  evaluating = signal<boolean>(false);
  evalError = signal<string | null>(null);

  constructor(private flagService: FlagService) {}

  test() {
    this.evalError.set(null);
    let attrs = {};
    try {
      attrs = JSON.parse(this.testAttributes());
    } catch (e) {
      this.evalError.set('Invalid JSON in attributes');
      return;
    }

    const context = { ...this.testContext(), attributes: attrs };
    this.evaluating.set(true);
    
    this.flagService.evaluate(this.flagKey(), context).subscribe({
      next: (res) => {
        this.testResult.set(res);
        this.evaluating.set(false);
      },
      error: (err) => {
        this.evalError.set(err.error?.error || 'Evaluation failed');
        this.evaluating.set(false);
      }
    });
  }

  updateContext(field: keyof EvaluationContext, value: string) {
    this.testContext.update(ctx => ({ ...ctx, [field]: value }));
  }
}
