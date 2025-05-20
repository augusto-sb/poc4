import { CanActivateFn } from '@angular/router';
import { inject } from '@angular/core';

import { SessionService } from './session.service';

export const generalGuard: CanActivateFn = (route, state) => {
  console.log(route, state);
  const session = inject(SessionService);
  return session.isLoggedIn();
};
