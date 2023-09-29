package auth

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/auth/jwtgenerator"
	"github.com/erupshis/bonusbridge/internal/auth/users"
	"github.com/erupshis/bonusbridge/internal/controllers"
	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/go-chi/chi/v5"
)

type controller struct {
	usersStrg users.Storage
	jwt       jwtgenerator.JwtGenerator

	log logger.BaseLogger
}

func CreateAuthenticator(usersStorage users.Storage, jwt jwtgenerator.JwtGenerator, baseLogger logger.BaseLogger) controllers.BaseController {
	return &controller{
		usersStrg: usersStorage,
		jwt:       jwt,
		log:       baseLogger,
	}
}

func (c *controller) Route() *chi.Mux {
	r := chi.NewRouter()

	r.Post("/register", c.registerHandler)

	r.Get("/login", c.loginHandlerForm)
	r.Post("/login", c.loginHandler)

	return r
}

func (c *controller) registerHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		c.log.Info("[controller:registerHandler] failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	var user users.User
	if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.log.Info("[controller:registerHandler] bad new user input data")
		return
	}

	userID, err := c.usersStrg.GetUserId(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:registerHandler] failed to check user in database")
		return
	}

	if userID != -1 {
		w.WriteHeader(http.StatusConflict)
		c.log.Info("[controller:registerHandler] login already exists")
		return
	}

	userID, err = c.usersStrg.AddUser(user.Login, user.Password)
	if err != nil || userID == -1 {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:registerHandler] failed to add new user '%s'", user.Login)
		return
	}

	token, err := c.jwt.BuildJWTString(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] new token generation failed: %w", err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusCreated)

	c.log.Info("[controller:registerHandler] user '%s' registered successfully", user.Login)
}

func (c *controller) loginHandler(w http.ResponseWriter, r *http.Request) {
	buf := bytes.Buffer{}
	if _, err := buf.ReadFrom(r.Body); err != nil {
		c.log.Info("[controller:loginHandler] failed to read request body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer helpers.ExecuteWithLogError(r.Body.Close, c.log)

	var user users.User
	if err := helpers.UnmarshalData(buf.Bytes(), &user); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		c.log.Info("[controller:loginHandler] bad new user input data")
		return
	}

	userID, err := c.usersStrg.GetUserId(user.Login)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] failed to get userID from user's database: %w", err)
		return
	}

	if userID == -1 {
		w.WriteHeader(http.StatusUnauthorized)
		c.log.Info("[controller:loginHandler] failed to get userID from user's database: %w", err)
		return
	}

	authorized, err := c.usersStrg.ValidateUser(user.Login, user.Password)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] failed to check user's login/password in database")
		return
	}

	if !authorized {
		w.WriteHeader(http.StatusUnauthorized)
		c.log.Info("[controller:loginHandler] failed to authorize user")
		return
	}

	token, err := c.jwt.BuildJWTString(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		c.log.Info("[controller:loginHandler] new token generation failed: %w", err)
		return
	}

	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)

	c.log.Info("[controller:registerHandler] user '%s' authenticated successfully", user.Login)
}

func (c *controller) loginHandlerForm(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Login</title>
		</head>
		<body>
			<h1>Login</h1>
			<form id="login-form">
				<label for="username">Username:</label>
				<input type="text" id="username" name="username" required><br>
				<label for="password">Password:</label>
				<input type="password" id="password" name="password" required><br>
				<input type="submit" value="Login">
			</form>
		
			<script>
				document.getElementById('login-form').addEventListener('submit', function(event) {
					event.preventDefault(); // Prevent the form from submitting normally
		
					// Get the form data
					const username = document.getElementById('username').value;
					const password = document.getElementById('password').value;
		
					// Create a JavaScript object
					const data = {
						login: username,
						password: password
					};
		
					// Convert the JavaScript object to JSON
					const jsonData = JSON.stringify(data);
		
					// You can send the JSON data in the request body
					fetch('/api/user/login', {
						method: 'POST',
						body: jsonData,
						headers: {
							'Content-Type': 'application/json'
						}
					})
					.then(response => {
						if (response.ok) {
							const authorizationHeader = response.headers.get("Authorization");
							if (authorizationHeader) {
           				 		const token = authorizationHeader.split(' ')[1];

            					if (token) {
                					alert("Login successful\nBearer Token: " + token);
            					} else {
                					alert('Token extraction failed');
            					}
        					} else {
            					alert("Authorization header missing in response");
							}
						} else {
							if (response.status === 401) {
								alert('Unauthorized: Please check your credentials');
        					} else {
								alert("Error: Status Code " + response.status);
        					}
						}
					})
					.catch(error => {
						console.error("Request error:", error);
						alert("Network or request error occurred");
					});
				});
			</script>
		</body>
		</html>
	`)
}
