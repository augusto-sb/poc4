import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Router } from '@angular/router';

@Injectable({
  providedIn: 'root',
})
export class SessionService {
  private loggedIn = false;

  constructor(
    private readonly httpClient: HttpClient,
    private readonly router: Router,
  ) {}

  public init(): Promise<void> {
    return new Promise((resolve, reject) => {
      this.httpClient.get<number>('/session', { responseType: 'json' }).subscribe(
        //resolve,
        x => {
          this.loggedIn = x !== 0;
          resolve();
        },
        reject,
        //err => {console.log(err);},
        () => {
          console.log('session init finish');
        },
      );
    });
  }

  public isLoggedIn(): boolean {
    return this.loggedIn;
  }

  /*
  public isLoggedIn(): boolean {
    return false;//(Math.random() > 0.5);
  }
  */

  public login(data: Partial<{ username: string | null; password: string | null }>): void {
    this.httpClient
      .post('/login', null, {
        responseType: 'text',
        headers: { Authorization: 'Basic ' + btoa(data.username + ':' + data.password) },
      })
      .subscribe(
        resp => {
          console.log(resp);
          this.loggedIn = true;
          this.router.navigate(['']);
        },
        err => {
          console.log(err);
          if (err.status === 401) {
            alert('usuario o contrasena incorrectos');
          } else {
            alert('error desconocido');
          }
        },
        () => {
          console.log('login finish');
        },
      );
  }

  public logout(): void {
    sessionStorage.clear();
    localStorage.clear();
    console.log(
      document.cookie.split('; ').map(x => {
        const tmp = x.split('=');
        return { k: tmp[0], v: tmp[1] };
      }),
    );
    //document.cookie = '';
    this.httpClient.post('/logout', null, undefined).subscribe(
      resp => {
        console.log(resp);
        this.loggedIn = false;
        this.router.navigate(['']);
      },
      console.log,
      () => {
        console.log('logout finish');
      },
    );
    return;
  }
}
