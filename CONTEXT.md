# go-react-template

This repo is a template for a Go and React application. Its example bounded context manages Users and their Registration.

## Language

**User**:
A person who has registered with the application.
_Avoid_: account, member, customer

**Email address**:
The address that identifies a User. Two addresses that are different only in letter case are the same address.
_Avoid_: login, username

**Display name**:
The name that the application shows for a User. Two Users can have the same display name.
_Avoid_: username, full name

**Registration**:
The creation of a User from an email address and a display name.
_Avoid_: sign-up, onboarding, create-user

**Welcome email**:
The one email that the application sends to a User after Registration.
_Avoid_: greeting, onboarding email
