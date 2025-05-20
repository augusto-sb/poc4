import { HttpInterceptorFn } from '@angular/common/http';
import { environment } from '../environments/environment';

export const apiInterceptor: HttpInterceptorFn = (req, next) => {
  const apiReq = req.clone({
    url: environment.apiUrl + req.url,
    withCredentials: true, // para la cookie
  });
  return next(apiReq);
};
