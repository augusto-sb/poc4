import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';

import 'reflect-metadata';

class Class {
  _path = '';
  constructor() {
    throw new Error('use a class!');
  }
}

@Component({
  selector: 'app-generic',
  imports: [],
  templateUrl: './generic.component.html',
  styleUrl: './generic.component.css',
})
export class GenericComponent<TData extends Class = Class> implements OnInit {
  public entities: TData[] = [];
  public keys: (keyof TData)[] = [];
  public count = 0;

  constructor(private readonly httpClient: HttpClient) {
    //console.log(arguments, Reflect.getMetadata, new TData())
    console.log(this);
  }

  ngOnInit(): void {
    this.httpClient
      .get<{ Results: TData[]; Count: number }>('/entities' /*AAAAA*/, { responseType: 'json' })
      .subscribe(
        resp => {
          if (this.entities.length) {
            this.keys = Object.keys(this.entities[0] as keyof TData) as (keyof TData)[];
          }
          this.entities = resp.Results;
          this.count = resp.Count;
        },
        err => {
          console.log(err);
          alert();
        },
        () => {
          console.log('finish');
        },
      );
    return;
  }
}
