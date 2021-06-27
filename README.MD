# apierr

This lib permits to create APIErrors avoiding to manage the error on the api side. Supports error check and error unwrapping.

Compatible with [go-kit](https://github.com/go-kit/kit).

### Example of usage
```
var errInvalidAge = fmt.Errorf("invalid age")
var errSameAge = fmt.Errorf("same age")
func (svc *UserService) UpdateUserAge(id, age int) error {
  user, err := svc.repository.FindById(id)
  // if err != nil ...
  if user.age == age {
    return apierr.New(errSameAge, http.StatusConflict)
  }
  // and so on...
}
```
### And on the Api resource

```
func (api *UserApi) HandleChangeAge(w http.ResponseWriter, r *http.Request) {
  // retrieve id, age...
  if err := api.userService.UpdateUserAge(id, age); err != nil {
    apierr.HandleISE(err, w)
    return
  }
}
```

The `HandleChangeAge` will write `409 Conflict` on the response.

Check the godoc for more information about usage/functionalities as Custom headers, decorators, ...