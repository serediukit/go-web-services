package main

import "encoding/binary"
import "bytes"

func (in *ApiError) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *MyApi) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *ProfileParams) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *CreateParams) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *User) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *NewUser) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *OtherApi) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *OtherCreateParams) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}

func (in *OtherUser) Unpack(data []byte) error {
	r := bytes.NewReader(data)
	return nil
}
