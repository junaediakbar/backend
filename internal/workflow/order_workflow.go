package workflow

import (
	"strings"

	"laundry-backend/internal/model"
)

// Tahap workflow (nilai orders.workflow_status)
const (
	Received    = "received"
	RontokDone  = "rontok_done"
	JemurDone   = "jemur_done"
	DownyDone   = "downy_done"
	PackingDone = "packing_done"
	Delivered   = "delivered"
	PickedUp    = "picked_up"
)

// Urutan produksi per item. Index 0 = rontok (opsional).
var ProductionChain = []struct {
	Canonical string
	Optional  bool
}{
	{"rontok", true},
	{"sikat", false},
	{"bilas", false},
	{"jemur", true}, // tidak lagi di template UI; data lama / alias spin_dry tetap valid
	{"downy", false},
	{"rumbai", true},
	{"finishing_1", false},
	{"finishing_2", false},
}

// Alias task lama → canonical
var TaskAliases = map[string]string{
	"dust_removal":  "rontok",
	"brushing":      "sikat",
	"rinse_sprayer": "bilas",
	"spin_dry":      "jemur",
	"spin_dry_1":    "jemur",
	"spin_dry_2":    "jemur",
	"finishing_packing": "finishing_1", // data lama: diperlakukan seperti Finishing 1 untuk rantai
}

// NormalizeTask maps task_type ke canonical key
func NormalizeTask(taskType string) string {
	t := strings.TrimSpace(taskType)
	if t == "" {
		return ""
	}
	if a, ok := TaskAliases[t]; ok {
		return a
	}
	return t
}

// HasAssignment true jika task punya karyawan terisi
func HasAssignment(was []model.WorkAssignment, canonical string) bool {
	c := strings.TrimSpace(canonical)
	for _, wa := range was {
		if NormalizeTask(wa.TaskType) == c && strings.TrimSpace(wa.Employee.ID) != "" {
			return true
		}
	}
	return false
}

// CanAssignTask: tugas berantai — semua tugas non-opsional sebelum target harus sudah diisi
func CanAssignTask(itemAssignments []model.WorkAssignment, canonicalTask string) bool {
	target := NormalizeTask(canonicalTask)
	if target == "" {
		return false
	}
	idx := findProductionIndex(target)
	if idx < 0 {
		return true // jemput/antar / bukan rantai produksi
	}
	for j := 0; j < idx; j++ {
		prev := ProductionChain[j]
		if prev.Optional {
			continue
		}
		if !HasAssignment(itemAssignments, prev.Canonical) {
			return false
		}
	}
	return true
}

func findProductionIndex(canonical string) int {
	for i, step := range ProductionChain {
		if step.Canonical == canonical {
			return i
		}
	}
	return -1
}

// Rank untuk perbandingan (lebih besar = lebih maju)
func Rank(status string) int {
	switch strings.TrimSpace(status) {
	case Received:
		return 0
	case RontokDone:
		return 10
	case JemurDone:
		return 20
	case DownyDone:
		return 30
	case PackingDone:
		return 40
	case Delivered:
		return 50
	case PickedUp:
		return 60
	case "washing":
		return 10
	case "drying":
		return 20
	case "ironing":
		return 30
	case "finished":
		return 40
	default:
		return 0
	}
}

// itemProductionMilestone: seberapa jauh satu sub-nota dalam tahap produksi (0 = belum mulai).
// Selaras dengan label status: rontok→dirontok, bilas/jemur→dijemur, downy→didowny, finishing_1→packing.
func itemProductionMilestone(was []model.WorkAssignment) int {
	if HasAssignment(was, "finishing_1") || HasAssignment(was, "finishing_packing") ||
		HasAssignment(was, "finishing_2") {
		return 4
	}
	if HasAssignment(was, "downy") {
		return 3
	}
	if HasAssignment(was, "bilas") || HasAssignment(was, "jemur") {
		return 2
	}
	if HasAssignment(was, "rontok") {
		return 1
	}
	return 0
}

// dropoffComplete — driver + bensin + buruh 1 + buruh 2 (total pembagian 7,5%)
func dropoffComplete(was []model.WorkAssignment) bool {
	has := func(keys ...string) bool {
		for _, k := range keys {
			if HasAssignment(was, k) {
				return true
			}
		}
		return false
	}
	hasDriver := has("dropoff_driver")
	hasFuel := has("dropoff_bensin", "dropoff_fuel")
	hasW1 := has("dropoff_buruh_1", "dropoff_worker_1")
	hasW2 := has("dropoff_buruh_2", "dropoff_worker_2")
	// Empat slot antar (total 7,5% terbagi); tanpa baris terpisah "antar/jemput"
	return hasDriver && hasFuel && hasW1 && hasW2
}

// TargetFromAssignments menghitung status dari performa karyawan:
// milestone terendah antar sub-nota menentukan status (rontok→dirontok, bilas/jemur→dijemur, downy→didowny, finishing_1→packing).
// Diantar setelah semua sub-nota selesai dropoff.
func TargetFromAssignments(o *model.OrderDetail) string {
	if o == nil || len(o.Items) == 0 {
		return Received
	}
	all := func(fn func([]model.WorkAssignment) bool) bool {
		for _, it := range o.Items {
			if !fn(it.WorkAssignments) {
				return false
			}
		}
		return true
	}
	minM := 4
	for _, it := range o.Items {
		m := itemProductionMilestone(it.WorkAssignments)
		if m < minM {
			minM = m
		}
	}
	if minM == 0 {
		return Received
	}
	switch minM {
	case 1:
		return RontokDone
	case 2:
		return JemurDone
	case 3:
		return DownyDone
	case 4:
		if all(dropoffComplete) {
			return Delivered
		}
		return PackingDone
	default:
		return Received
	}
}
