import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import {BrowserAnimationsModule} from '@angular/platform-browser/animations';



import { AppComponent } from './app.component';
import { AppRoutingModule } from './/app-routing.module';
import { MaterialImportModule } from './/material-import.module';
import { TeacherService } from './teacher.service'

@NgModule({
  declarations: [
    AppComponent
  ],
  imports: [
    BrowserModule,
    AppRoutingModule,
    BrowserAnimationsModule,
    MaterialImportModule,
  ],
  providers: [TeacherService],
  bootstrap: [AppComponent]
})
export class AppModule { }