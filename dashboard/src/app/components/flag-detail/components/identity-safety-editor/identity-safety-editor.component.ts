import { Component, Input } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule, FormGroup } from '@angular/forms';

@Component({
  selector: 'app-identity-safety-editor',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule],
  templateUrl: './identity-safety-editor.component.html',
  styleUrl: './identity-safety-editor.component.css'
})
export class IdentitySafetyEditorComponent {
  @Input({ required: true }) form!: FormGroup;
}
