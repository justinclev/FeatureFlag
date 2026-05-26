import { Routes } from '@angular/router';
import { FlagListComponent } from './components/flag-list/flag-list.component';
import { FlagDetailComponent } from './components/flag-detail/flag-detail.component';

export const routes: Routes = [
  { path: '', component: FlagListComponent },
  { path: 'flags', component: FlagListComponent },
  { path: 'flags/new', component: FlagDetailComponent },
  { path: 'flags/:id', component: FlagDetailComponent },
];
