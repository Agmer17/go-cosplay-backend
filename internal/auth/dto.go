package auth

type googleUserResponse struct {
	ID            string `json:"id"`             // ID unik Google (ProviderUserID)
	Email         string `json:"email"`          // Alamat email
	VerifiedEmail bool   `json:"verified_email"` // Status verifikasi email
	Name          string `json:"name"`           // Nama lengkap
	GivenName     string `json:"given_name"`     // Nama depan
	FamilyName    string `json:"family_name"`    // Nama belakang
	Picture       string `json:"picture"`        // URL foto profil
	Locale        string `json:"locale"`         // Preferensi bahasa (misal: "id" atau "en")
}
