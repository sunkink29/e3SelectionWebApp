import { Component, OnInit } from '@angular/core';
import { Observable } from 'rxjs/Observable';
import { ActivatedRoute, Router, ParamMap } from '@angular/router';
import { MatDialog, MatDialogRef } from '@angular/material';
import { map, filter } from 'rxjs/operators';


import { Teacher, Block} from '../../Interfaces';
import { TeacherService } from '../../teacher.service';
import { ChangeDialogComponent } from '../change-dialog/change-dialog.component';

@Component({
  selector: 'app-student-dashboard',
  templateUrl: './student-dashboard.component.html',
  styleUrls: ['./student-dashboard.component.css']
})
export class StudentDashboardComponent implements OnInit {

  index = 0
  curDisplayedColumns = ['days', 'name', 'roomNumber', 'status', 'subject', 'description', 'size'];
  nextDisplayedColumns = ['days', 'name', 'roomNumber', 'status', 'subject', 'description', 'size', 'changeButton'];

  curTeachers$: Observable<Teacher[]>;
  nextTeachers$: Observable<Teacher[]>;

  blockID: number;

  constructor(
    private teacherService: TeacherService, 
    private route: ActivatedRoute,
    private router: Router,
    public dialog: MatDialog) { }

  ngOnInit() {
    this.curTeachers$ = this.teacherService.getCurClasses();
    this.nextTeachers$ = this.teacherService.getNextClasses();
    let dialogRef;
    setTimeout(() => {this.route.url
      .subscribe(url => {
        let index = url.findIndex((value) => value.toString() == 'change')
        if (index != -1) {
          dialogRef = this.dialog.open(ChangeDialogComponent, {
            height: '70%',
            width: '80%',
            data: this.blockID
          });
          dialogRef.afterClosed().subscribe(email => {
            if (email != null) {
              this.teacherService.setStudentClass(email, this.blockID);
            }
            this.router.navigate(['/student']);
          });
        }
      });
    },);
    this.route.paramMap.pipe(map((params: ParamMap) => params.get('id'))).subscribe(
      id => {this.blockID = +id}, () => {})
  }

  getIndex(increase: boolean) {
    if (increase) {
      this.index++;
    }
    return this.index % 2;
  }
}