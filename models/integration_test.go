//go:build integration

package models

// run test with this command: go test . --tags integration --count=1

// these integration test start a docker image with postgres, then they add tables and perform
// CRUD operations on them
import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgconn"
	_ "github.com/jackc/pgx/v4"
	_ "github.com/jackc/pgx/v4/stdlib"
	dockertest "github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
)

var (
	host     = "localhost"
	user     = "imperator"
	password = "password"
	dbName   = "imperator_test"
	port     = "5435"
	dsn      = "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable timezone=UTC connect_timeout=5"
)

var dummyUser = User{
	FirstName: "John",
	LastName:  "Smith",
	Email:     "test@example.com",
	Active:    1,
	Password:  "password",
}

var models Models
var testDB *sql.DB
var resource *dockertest.Resource
var pool *dockertest.Pool

func TestMain(m *testing.M) {
	fmt.Println("TestMain...")

	os.Setenv("DATABASE_TYPE", "postgres")
	os.Setenv("UPPER_DB_LOG", "ERROR")

	var err error
	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker")
	}

	opts := dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "13.4",
		Env: []string{
			"POSTGRES_USER=" + user,
			"POSTGRES_PASSWORD=" + password,
			"POSTGRES_DB=" + dbName,
		},
		ExposedPorts: []string{"5432"},
		PortBindings: map[docker.Port][]docker.PortBinding{
			"5432": {
				{HostIP: "0.0.0.0", HostPort: port},
			},
		},
	}

	resource, err = pool.RunWithOptions(&opts)
	if err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not start resource with error: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		testDB, err = sql.Open("pgx", fmt.Sprintf(dsn, host, port, user, password, dbName))
		if err != nil {
			return err
		}
		return testDB.Ping()
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to docker with error: %s", err)
	}

	// docker image has started with postgres
	err = createTables(testDB)
	if err != nil {
		log.Fatalf("error createing the tables: %s", err)
	}

	models = New(testDB)

	code := m.Run()

	// for debugging the db for integration test we can comment out these lines to kill
	// the instance below, once done remove the comments
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("could not purge resource: %s", err)
	}

	os.Exit(code)
}

func createTables(db *sql.DB) error {
	var stmt = `
      
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

drop table if exists users cascade;

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name character varying(255) NOT NULL,
    last_name character varying(255) NOT NULL,
    user_active integer NOT NULL DEFAULT 0,
    email character varying(255) NOT NULL UNIQUE,
    password character varying(60) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

drop table if exists remember_tokens;

CREATE TABLE remember_tokens (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    remember_token character varying(100) NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone NOT NULL DEFAULT now()
);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON remember_tokens
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

drop table if exists tokens;

CREATE TABLE tokens (
    id SERIAL PRIMARY KEY,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    first_name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    token character varying(255) NOT NULL,
    token_hash bytea NOT NULL,
    created_at timestamp without time zone NOT NULL DEFAULT now(),
    updated_at timestamp without time zone NOT NULL DEFAULT now(),
    expiry timestamp without time zone NOT NULL
);

CREATE TRIGGER set_timestamp
    BEFORE UPDATE ON tokens
    FOR EACH ROW
    EXECUTE PROCEDURE trigger_set_timestamp();

`
	_, err := db.Exec(stmt)
	if err != nil {
		return err
	}
	return nil
}

func TestUser_Table(t *testing.T) {
	fmt.Println("TestUser_Table...")
	s := models.Users.Table()
	if s != "users" {
		t.Error("wrong table name returned", s)
	}
}

func TestUser_Insert(t *testing.T) {
	fmt.Println("TestUser_Insert...")
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("failed to insert user", err)
	}
	if id == 0 {
		t.Error("error got user with id 0 after insert")
	}
}

func TestUser_Get(t *testing.T) {
	fmt.Println("TestUser_Get...")
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}
	if u.ID == 0 {
		t.Error("id of returned user is 0:", err)
	}
}

func TestUser_GetAll(t *testing.T) {
	fmt.Println("TestUser_GetAll...")
	_, err := models.Users.GetAll()
	if err != nil {
		t.Error("failed to get sll users: ", err)
	}
}

func TestUser_GetByEmail(t *testing.T) {
	fmt.Println("TestUser_GetByEmail...")
	u, err := models.Users.GetByEmail("test@example.com")
	if err != nil {
		t.Error("failed to get users by email:", err)
	}
	if u.ID == 0 {
		t.Error("id of returned user is 0:", err)
	}
}

func TestUser_Update(t *testing.T) {
	fmt.Println("TestUser_Update...")
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	u.LastName = "Jackson"
	err = u.Update(*u)
	if err != nil {
		t.Error("failed to update the user:", err)
	}

	u, err = models.Users.Get(1)
	if err != nil {
		t.Error("failed to get updated user:", err)
	}

	if u.LastName != "Jackson" {
		t.Error("failed to update the last name in the database")
	}
}

func TestUser_PasswordMatches(t *testing.T) {
	fmt.Println("TestUser_PasswordMatches...")
	u, err := models.Users.Get(1)
	if err != nil {
		t.Error("failed to get user:", err)
	}

	matches, err := u.PasswordMatches("password")
	if err != nil {
		t.Error("error checking match:", err)
	}
	if !matches {
		t.Error("password does not match when it should")
	}

	matches, err = u.PasswordMatches("123")
	if err != nil {
		t.Error("error checking match:", err)
	}
	if matches {
		t.Error("password match when it should not")
	}
}

func TestUser_ResetPassword(t *testing.T) {
	fmt.Println("TestUser_ResetPassword...")
	err := models.Users.ResetPassword(1, "new_password")
	if err != nil {
		t.Error("error resetting password:", err)
	}

	err = models.Users.ResetPassword(2, "new_password")
	if err == nil {
		t.Error("did not get an error when trying to reset the password for a non existing user")
	}
}

func TestUser_Delete(t *testing.T) {
	fmt.Println("TestUser_Delete...")
	err := models.Users.Delete(1)
	if err != nil {
		t.Error("failed to delete the user:", err)
	}

	_, err = models.Users.Get(1)
	if err == nil {
		t.Error("retrieved user who was supposed to be deleted")
	}
}

func TestToken_Table(t *testing.T) {
	fmt.Println("TestToken_Table...")
	s := models.Tokens.Table()
	if s != "tokens" {
		t.Error("wrong table name returned, expected tokens and got:", s)
	}
}

func TestToken_GenerateToken(t *testing.T) {
	fmt.Println("TestToken_GenerateToken...")
	id, err := models.Users.Insert(dummyUser)
	if err != nil {
		t.Error("error inserting user:", err)
	}

	_, err = models.Tokens.GenerateToken(id, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token:", err)
	}
}

func TestToken_Insert(t *testing.T) {
	fmt.Println("TestToken_Insert...")
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get the user:", err)
	}

	token, err := models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("error generating token:", err)
	}

	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("error inserting token:", err)
	}
}

func TestToken_GetUserForToken(t *testing.T) {
	fmt.Println("TestToken_GetUserForToken...")
	fakeToken := "abc"
	_, err := models.Tokens.GetUserForToken(fakeToken)
	if err == nil {
		t.Error("error expected but not received when getting user with bad token")
	}

	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get the user:", err)
	}

	_, err = models.Tokens.GetUserForToken(u.Token.PlainText)
	if err != nil {
		t.Error("failed to get user with valid token:", err)
	}
}

func TestToken_GetTokensForUser(t *testing.T) {
	fmt.Println("TestToken_GetUserForToken...")
	tokens, err := models.Tokens.GetTokensForUser(1)
	if err != nil {
		t.Error("failed to get tokens for the user:", err)
	}

	if len(tokens) > 0 {
		t.Error("tokens returned for non-existing user")
	}
}

func TestToken_Get(t *testing.T) {
	fmt.Println("TestToken_Get...")
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get the user:", err)
	}

	_, err = models.Tokens.Get(u.Token.ID)
	if err != nil {
		t.Error("error getting token by id:", err)
	}
}

func TestToken_GetByToken(t *testing.T) {
	fmt.Println("TestToken_GetByToken...")
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get the user:", err)
	}

	_, err = models.Tokens.GetByToken(u.Token.PlainText)
	if err != nil {
		t.Error("error getting token by token:", err)
	}

	_, err = models.Tokens.GetByToken("fake")
	if err == nil {
		t.Error("no error getting non-existing token")
	}
}

var authData = []struct {
	name          string
	token         string
	email         string
	errorExpected bool
	message       string
}{
	{"invalid", "abcdefghijklmnopqrstuvwxyz", "fake@example.com", true, "invalid token accepted as valid"},
	{"invalid_length", "abcdefghijklmnopqrstuvwxyz", "fake@example.com", true, "token of wrong length"},
	{"no_user", "abcdefghijklmnopqrstuvwxyz", "fake@example.com", true, "no user but token accepted as valid"},
	{"valid", "", "test@example.com", false, "valid token reported as invalid"},
}

func TestToken_AuthenticateToken(t *testing.T) {
	fmt.Println("TestToken_AuthenticateToken...")
	for _, tt := range authData {
		token := ""
		if tt.email == dummyUser.Email {
			user, err := models.Users.GetByEmail(tt.email)
			if err != nil {
				t.Error("failed to get user:", err)
			}
			token = user.Token.PlainText
		} else {
			token = tt.token
		}

		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Add("Authorization", "Bearer "+token)

		_, err := models.Tokens.AuthenticationToken(req)
		if tt.errorExpected && err == nil {
			t.Errorf("%s: %s", tt.name, tt.message)
		} else if !tt.errorExpected && err != nil {
			t.Errorf("%s: %s - %s", tt.name, tt.message, err)
		} else {
			t.Logf("passed: %s", tt.name)
		}
	}
}

func TestToken_Delete(t *testing.T) {
	fmt.Println("TestToken_Delete...")
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("error getting the user:", err)
	}

	err = models.Tokens.DeleteByToken(u.Token.PlainText)
	if err != nil {
		t.Error("error deleting token:", err)
	}
}

func TestToken_ExpiredToken(t *testing.T) {
	fmt.Println("TestToken_ExpiredToken...")
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("error getting user:", err)
	}

	// notice the negative time here to make the token expired (-)time.Hour
	token, err := models.Tokens.GenerateToken(u.ID, -time.Hour*24*365)
	if err != nil {
		t.Error("error generating token:", err)
	}

	err = models.Tokens.Insert(*token, *u)
	if err != nil {
		t.Error("error inserting token:", err)
	}

	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token.PlainText)

	_, err = models.Tokens.AuthenticationToken(req)
	if err == nil {
		t.Error("failed to catch expired token:", err)
	}
}

func TestToken_BadHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	_, err := models.Tokens.AuthenticationToken(req)
	if err == nil {
		t.Error("failed to catch missing authentication header")
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "abc")

	_, err = models.Tokens.AuthenticationToken(req)
	if err == nil {
		t.Error("failed to catch bad authentication header")
	}

	newUser := User{
		FirstName: "first",
		LastName:  "last",
		Email:     "you@here.com",
		Active:    1,
		Password:  "abc",
	}

	id, err := models.Users.Insert(newUser)
	if err != nil {
		t.Error("failed to create a user:", err)
	}

	token, err := models.Tokens.GenerateToken(id, time.Hour*24*365)
	if err != nil {
		t.Error("failed to generate a token:", err)
	}

	err = models.Tokens.Insert(*token, newUser)
	if err != nil {
		t.Error("failed to insert token for user:", err)
	}

	err = models.Users.Delete(id)
	if err != nil {
		t.Error("failed to delte the user:", err)
	}

	req, _ = http.NewRequest("GET", "/", nil)
	req.Header.Add("Authorization", "Bearer "+token.PlainText)
	_, err = models.Tokens.AuthenticationToken(req)
	if err == nil {
		t.Error("failed to catch token for deleted user")
	}
}

func TestToken_DeletingNonExistingToken(t *testing.T) {
	err := models.Tokens.DeleteByToken("abc")
	if err != nil {
		t.Error("error deleting token:", err)
	}
}

func TestToken_ValidToken(t *testing.T) {
	u, err := models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get the user:", err)
	}

	newToken, err := models.Tokens.GenerateToken(u.ID, time.Hour*24*365)
	if err != nil {
		t.Error("failed to generate token:", err)
	}

	err = models.Tokens.Insert(*newToken, *u)
	if err != nil {
		t.Error("failed to insert token:", err)
	}

	ok, err := models.Tokens.ValidToken(newToken.PlainText)
	if err != nil {
		t.Error("error while calling token validation:", err)
	}
	if !ok {
		t.Error("valid token reported as invalid")
	}

	ok, _ = models.Tokens.ValidToken("invalid_token")
	if ok {
		t.Error("invalid token reported as valid")
	}

	u, err = models.Users.GetByEmail(dummyUser.Email)
	if err != nil {
		t.Error("failed to get the user:", err)
	}

	err = models.Tokens.Delete(u.Token.ID)
	if err != nil {
		t.Error("failed to delete token:", err)
	}

	ok, err = models.Tokens.ValidToken(u.Token.PlainText)
	if err == nil {
		t.Error("token is not valid:", err)
	}
	if ok {
		t.Error("no error reported when validation non-existing token")
	}
}
