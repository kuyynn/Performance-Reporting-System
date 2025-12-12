package model

type AdminCreateUserRequest struct {
	Username     string `json:"username"`
	FullName     string `json:"full_name"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	Role         string `json:"role"`

	// Mahasiswa
	ProgramStudy string `json:"program_study"`
	AcademicYear string `json:"academic_year"`
	AdvisorID    *int64 `json:"advisor_id"`

	// Dosen Wali
	NIP        string `json:"nip"`
	Department string `json:"department"`
}

