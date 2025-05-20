import { Routes } from '@angular/router';

import { AboutComponent } from './about/about.component';
import { HomeComponent } from './home/home.component';
import { LoginComponent } from './login/login.component';
import { EntitiesComponent } from './entities/entities.component';
//import { GenericComponent } from './generic/generic.component';
//import { Entity } from './entities/entity.class';
import { generalGuard } from './general.guard';

export const routes: Routes = [
  {
    path: '',
    component: HomeComponent,
  },
  {
    path: 'about',
    component: AboutComponent,
  },
  {
    path: 'entities',
    component: EntitiesComponent,
    //component: GenericComponent<Entity>,
    canActivate: [generalGuard],
  },
  {
    path: 'login',
    component: LoginComponent,
  },
];
