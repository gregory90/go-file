package model

import (
	"time"
)

type File struct {
	UID          string    `json:"uid,omitempty"`
	Name         string    `json:"name,omitempty"`
	ExternalName string    `json:"externalName,omitempty"`
	Mime         string    `json:"mime,omitempty"`
	UniqueID     string    `json:"uniqueID,omitempty"`
	Type         string    `json:"type,omitempty"`
	Tmp          bool      `json:"tmp,omitempty"`
	CreatedAt    time.Time `json:"createdAt,omitempty"`
}

func (m File) Validate() map[string]interface{} {
	//var errors binding.Errors

	//v := validation.NewValidation(&errors, m)
	//v.KeyTag("json")

	////v.Validate(&user.Email).Key("email").Message("required").Required()
	////v.Validate(&user.Email).Message("incorrect").Email()

	////v.Validate(&user.Password).Key("password").Message("required").Required()
	////v.Validate(&user.Password).Message("range").Range(6, 60)

	//return *v.Errors.(*binding.Errors)
	return nil
}
