# eti-assignment1-trips-backend

This is the second microservice of the ETI assignment 1. This backend handles everything related to trips. Please run the db first before starting this.

To run the backend, run `go run main.go` in the project command prompt.

## Project repos
* [Frontend](https://github.com/ngrayzin/eti-assignment1-frontend)
* [Db](https://github.com/ngrayzin/eti-assignment1-db)
* [User Backend](https://github.com/ngrayzin/eti-assignment1-user-backend)
* [Trips Backend](https://github.com/ngrayzin/eti-assignment1-trips-backend)

## Considerations for trips backend:
- For time conflicts for enrollment, passengers are not allowed to book a trip within a 30 min time range of a current booked trip.<br>
  Here is an example, I have a booking at 2pm.<br>
  Lets say I want to book a trip at 2:30pm. This is allowed ✅<br>
  I also want to book a trip at 1:30pm. This is allowed ✅<br>
  However, if I want to book a trip within the time range of 2pm to 2:29pm, this is not allowed ❎<br>
  If I also want to book a trip within 1:31pm to 2pm, this is not allowed ❎<br>

- Trips cannot be edited by car owners. The pick up location, alt pick up location and destination have no validation. We trust that car owners would put accurate and correct addresses and would pick people up at that correct time 😃

## Architecture diagram

![Blank diagram - Page 1](https://github.com/ngrayzin/eti-assignment1-trips-backend/assets/94064635/e3e80b00-647d-4e48-8cad-4ff6a061813e)
