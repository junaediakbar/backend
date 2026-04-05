package workflow

import (
	"testing"

	"github.com/stretchr/testify/require"

	"laundry-backend/internal/model"
)

func TestCanAssignTask_RontokOptional(t *testing.T) {
	t.Parallel()
	was := []model.WorkAssignment{}
	require.True(t, CanAssignTask(was, "sikat"))
	require.False(t, CanAssignTask(was, "bilas"))
}

func TestCanAssignTask_DownyAfterBilasWithoutJemur(t *testing.T) {
	t.Parallel()
	var a, b model.WorkAssignment
	a.TaskType, a.Employee.ID = "sikat", "e1"
	b.TaskType, b.Employee.ID = "bilas", "e1"
	was := []model.WorkAssignment{a, b}
	require.True(t, CanAssignTask(was, "downy"))
}

func TestCanAssignTask_FinishingAfterDownyWithoutRumbai(t *testing.T) {
	t.Parallel()
	var a, b, c model.WorkAssignment
	a.TaskType, a.Employee.ID = "sikat", "e1"
	b.TaskType, b.Employee.ID = "bilas", "e1"
	c.TaskType, c.Employee.ID = "downy", "e1"
	was := []model.WorkAssignment{a, b, c}
	require.True(t, CanAssignTask(was, "finishing_1"))
}

func TestCanAssignTask_Finishing2AfterFinishing1WithoutRumbai(t *testing.T) {
	t.Parallel()
	var a, b, c, d model.WorkAssignment
	a.TaskType, a.Employee.ID = "sikat", "e1"
	b.TaskType, b.Employee.ID = "bilas", "e1"
	c.TaskType, c.Employee.ID = "downy", "e1"
	d.TaskType, d.Employee.ID = "finishing_1", "e1"
	was := []model.WorkAssignment{a, b, c, d}
	require.True(t, CanAssignTask(was, "finishing_2"))
}

func TestCanAssignTask_FinishingPackingCountsAsFinishing1ForFinishing2(t *testing.T) {
	t.Parallel()
	var a, b, c, d model.WorkAssignment
	a.TaskType, a.Employee.ID = "sikat", "e1"
	b.TaskType, b.Employee.ID = "bilas", "e1"
	c.TaskType, c.Employee.ID = "downy", "e1"
	d.TaskType, d.Employee.ID = "finishing_packing", "e1"
	was := []model.WorkAssignment{a, b, c, d}
	require.True(t, CanAssignTask(was, "finishing_2"))
}

func TestTargetFromAssignments_Empty(t *testing.T) {
	t.Parallel()
	o := &model.OrderDetail{Items: []model.OrderItem{{ID: "i1", WorkAssignments: []model.WorkAssignment{}}}}
	require.Equal(t, Received, TargetFromAssignments(o))
}

func wa(task, emp string) model.WorkAssignment {
	var w model.WorkAssignment
	w.TaskType = task
	w.Employee.ID = emp
	return w
}

func TestTargetFromAssignments_Milestones(t *testing.T) {
	t.Parallel()
	o1 := func(was ...model.WorkAssignment) *model.OrderDetail {
		return &model.OrderDetail{Items: []model.OrderItem{{ID: "i1", WorkAssignments: was}}}
	}
	require.Equal(t, RontokDone, TargetFromAssignments(o1(wa("rontok", "e1"))))
	require.Equal(t, JemurDone, TargetFromAssignments(o1(wa("rontok", "e1"), wa("bilas", "e1"))))
	require.Equal(t, DownyDone, TargetFromAssignments(o1(
		wa("rontok", "e1"), wa("bilas", "e1"), wa("downy", "e1"),
	)))
	require.Equal(t, PackingDone, TargetFromAssignments(o1(
		wa("rontok", "e1"), wa("bilas", "e1"), wa("downy", "e1"), wa("finishing_1", "e1"),
	)))
}

func TestTargetFromAssignments_WeakestItem(t *testing.T) {
	t.Parallel()
	o := &model.OrderDetail{Items: []model.OrderItem{
		{ID: "a", WorkAssignments: []model.WorkAssignment{wa("rontok", "e1"), wa("bilas", "e1")}},
		{ID: "b", WorkAssignments: []model.WorkAssignment{wa("rontok", "e1")}},
	}}
	require.Equal(t, RontokDone, TargetFromAssignments(o))
}
