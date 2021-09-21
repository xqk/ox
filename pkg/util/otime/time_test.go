package otime

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestTime(t *testing.T) {
	now := Now().Truncate(time.Minute).Unix()
	fmt.Printf("now = %+v\n", now)
	fmt.Printf("time.Now() = %+v\n", time.Now().Unix())
	fmt.Println(Now().BeginOfDay().String())
	time.Sleep(time.Second)
}

func TestNow(t *testing.T) {
	type test struct {
		name string
		want *Time
	}
	var tests []*test
	tests = append(tests, &test{
		name: "2021-09-22 00:00:00",
		want: Unix(1632240000, 0),
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Now(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Now() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnix(t *testing.T)  {
	type args struct {
		sec  int64
		nsec int64
	}
	type test struct {
		name string
		args args
		want *Time
	}
	var tests []*test
	tests = append(tests, &test{
		name: "2021-09-22 00:00:00",
		args: args{
			sec: 1632240000,
			nsec: 0,
		},
		want: Unix(1632240000, 0),
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Unix(tt.args.sec, tt.args.nsec); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Unix() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToday(t *testing.T) {
	type test struct {
		name string
		want *Time
	}
	var tests []*test
	tests = append(tests, &test{
		name: "2021-09-21 23:59:59",
		want: Unix(1632239999, 0),
	})
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Today(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Today() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_BeginOfYear(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "beg of year",
			fields: fields{
				Time: time.Date(2021, 9, 13, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2021, 1, 1, 0, 0, 0, 0, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.BeginOfYear(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.BeginOfYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_EndOfYear(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "end of year",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 12, 31, 23, 59, 59, 999999999, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.EndOfYear(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.EndOfYear() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_BeginOfMonth(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "beg of month",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 1, 0, 0, 0, 0, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.BeginOfMonth(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.BeginOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_EndOfMonth(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "end of month",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 30, 23, 59, 59, 999999999, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.EndOfMonth(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.EndOfMonth() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_BeginOfWeek(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "beg of week",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 5, 31, 0, 0, 0, 0, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.BeginOfWeek(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.BeginOfWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_EndOfWeek(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "end of week",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 6, 23, 59, 59, 999999999, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.EndOfWeek(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.EndOfWeek() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_BeginOfDay(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "beg of day",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 3, 0, 0, 0, 0, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.BeginOfDay(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.BeginOfDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_EndOfDay(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "end of day",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 3, 23, 59, 59, 999999999, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.EndOfDay(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.EndOfDay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_BeginOfHour(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "begin of hour",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 3, 12, 0, 0, 0, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.BeginOfHour(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.BeginOfHour() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_EndOfHour(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "end of hour",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 3, 12, 59, 59, 999999999, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.EndOfHour(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.EndOfHour() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_BeginOfMinute(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "beg of minute",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 3, 12, 13, 0, 0, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.BeginOfMinute(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.BeginOfMinute() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTime_EndOfMinute(t *testing.T) {
	type fields struct {
		Time time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   *Time
	}{
		{
			name: "end of minute",
			fields: fields{
				Time: time.Date(2020, 6, 3, 12, 13, 11, 189, time.Local),
			},
			want: &Time{time.Date(2020, 6, 3, 12, 13, 59, 999999999, time.Local)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ti := &Time{
				Time: tt.fields.Time,
			}
			if got := ti.EndOfMinute(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Time.EndOfMinute() = %v, want %v", got, tt.want)
			}
		})
	}
}
