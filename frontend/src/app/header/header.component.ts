import { Component } from '@angular/core';
import { RouterLink } from '@angular/router';
import { SessionService } from '../session.service';

@Component({
  selector: 'app-header',
  imports: [RouterLink],
  templateUrl: './header.component.html',
  styleUrl: './header.component.css',
})
export class HeaderComponent {
  constructor(public readonly session: SessionService) {}

  public logout(): void {
    if (confirm('¿seguro?')) {
      this.session.logout();
    }
  }
}
