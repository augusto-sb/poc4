package main

// imports

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// types

type session struct {
	security  string
	timestamp int64
	info      map[string]any
}

type user struct {
	name     string
	password string
}

// consts

const cookieName string = "PHPSESSID" //gg
const cookieDurationInSeconds = 3600
const cleanerInterval = 30
const ctxKey string = "yourKey"

// vars

var muS sync.Mutex = sync.Mutex{}
var muU sync.Mutex = sync.Mutex{}
var sessions map[string]session = map[string]session{}
var users []user = []user{
	user{
		name:     "admin",
		password: "admin",
	},
}

// helpers

func genCookie(val string, delete bool) *http.Cookie {
	MaxAge := -1
	if !delete {
		MaxAge = cookieDurationInSeconds
	}
	return &http.Cookie{
		Name:     cookieName,
		Value:    val,
		MaxAge:   MaxAge,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}
}

func logger(msj string) {
	if os.Getenv("LOGGER") == "true" {
		fmt.Println(msj)
	}
}

/*func getSession(req *http.Request) *session{ // mmmmm
	cookie, err := req.Cookie(cookieName)
	if (err != nil){
		return nil;
	}
	muS.Lock();
	val, ok := sessions[cookie.Value];
	muS.Unlock();
	if(!ok){
		return nil;
	}
	return &val;
}*/

//gracefull shutdown implementar!!!

func setCleaner(timeSec uint) {
	//timer cada tanto limpia sessions vencidas!
	ticker := time.NewTicker(time.Duration(timeSec*1000) * time.Millisecond)
	done := make(chan bool)
	go func() {
		for {
			select {
			case <-ticker.C:
				logger("running cleaner!")
				muS.Lock()
				for k, v := range sessions {
					if v.timestamp+(cookieDurationInSeconds*1000) < time.Now().UnixMilli() {
						logger("cleaner found expired: " + k)
						delete(sessions, k)
					}
				}
				muS.Unlock()
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()
}

// middlewares

func sessionMiddleware(next http.HandlerFunc, protected bool) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var shouldSetCookie bool = false
		now := time.Now()
		unixMilli := now.UnixMilli()
		cookie, err := req.Cookie(cookieName)
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) {
				shouldSetCookie = true
			} else {
				logger("error retrieving cookie")
				http.Error(rw, "server error", http.StatusInternalServerError)
				return
			}
		}
		//get info de origen para que no me la puedan usar de otro lado!
		//el proxy deberia pasarme el header o algo de info! como ip publica etc
		//logger(req.Header.Values("User-Agent")) //recibir en x-forwarded-for por ej! para bien seguro
		pseudoSecure := req.Header.Get("User-Agent") + req.Header.Get("Accept") + req.Header.Get("Host") + req.Header.Get("X-Forwarded-For") + req.Header.Get("Forwarded")
		logger(pseudoSecure)
		if !shouldSetCookie {
			logger(cookie.Value)
			//check validity
			muS.Lock()
			val, ok := sessions[cookie.Value]
			muS.Unlock()
			if ok {
				if val.timestamp+(cookieDurationInSeconds*1000) < unixMilli {
					// expired
					muS.Lock()
					delete(sessions, cookie.Value)
					muS.Unlock()
					shouldSetCookie = true
				} else {
					//actualizar timestamp
					val.timestamp = unixMilli
					muS.Lock()
					sessions[cookie.Value] = val
					muS.Unlock()
					cookie.MaxAge = cookieDurationInSeconds
				}
				if val.security != pseudoSecure {
					muS.Lock()
					delete(sessions, cookie.Value)
					muS.Unlock()
					rw.WriteHeader(http.StatusNotAcceptable)
					rw.Write([]byte("hacker wtf\n"))
					return
				}
				if len(val.info) == 0 {
					if protected {
						// protected handler and not logged in - future use of permissions
						rw.WriteHeader(http.StatusUnauthorized)
						return
					}
				}
			} else {
				//rw.Write([]byte("cookie must have expired found in session!\n"))
				shouldSetCookie = true
			}
		}
		if shouldSetCookie {
			h := md5.New()
			io.WriteString(h, now.String())
			io.WriteString(h, pseudoSecure)
			cookieValueAndSessionKey := hex.EncodeToString(h.Sum(nil))
			cookie = genCookie(cookieValueAndSessionKey, false)
			newSession := session{
				timestamp: unixMilli,
				security:  pseudoSecure,
			}
			muS.Lock()
			sessions[cookieValueAndSessionKey] = newSession
			muS.Unlock()
		}
		http.SetCookie(rw, cookie)
		ctx := context.WithValue(req.Context(), ctxKey, cookie.Value)
		req = req.WithContext(ctx)
		next.ServeHTTP(rw, req)
	})
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	var corsOrigin string = os.Getenv("CORS_ORIGIN")
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if corsOrigin != "" {
			rw.Header().Set("Access-Control-Allow-Credentials", "true")
			rw.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
			rw.Header().Set("Access-Control-Allow-Headers", "authorization")
			rw.Header().Set("Access-Control-Allow-Methods", "GET,POST")
		}
		if req.Method == http.MethodOptions {
			return
		}
		next.ServeHTTP(rw, req)
	})
}

func middleware(next http.HandlerFunc, protected bool) http.HandlerFunc {
	// chain de todos los middlewares
	return corsMiddleware(sessionMiddleware(next, protected))
}

// handlers

func sessionHandler(rw http.ResponseWriter, req *http.Request) {
	/*var sess *session = getSession(req);
	if(sess == nil){
		rw.WriteHeader(http.StatusInternalServerError);
		return;
	}*/
	cookieValueAndSessionKey, ok := req.Context().Value(ctxKey).(string)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	muS.Lock()
	sess, ok := sessions[cookieValueAndSessionKey]
	muS.Unlock()
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	byteArr, err := json.Marshal(len(sess.info))
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Write(byteArr)
}

func loginHandler(rw http.ResponseWriter, req *http.Request) {
	cookieValueAndSessionKey, ok := req.Context().Value(ctxKey).(string)
	logger(cookieValueAndSessionKey)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	u, p, ok := req.BasicAuth()
	if !ok {
		rw.WriteHeader(http.StatusBadRequest)
		return
	}
	validUserPass := false
	muU.Lock()
	for _, val := range users {
		if val.name == u && val.password == p {
			validUserPass = true
			break
		}
	}
	muU.Unlock()
	if validUserPass {
		muS.Lock()
		sess, ok := sessions[cookieValueAndSessionKey]
		if !ok {
			rw.WriteHeader(http.StatusBadRequest)
		} else {
			sess.info = map[string]any{"asd": "asd"}
			sessions[cookieValueAndSessionKey] = sess
			rw.WriteHeader(http.StatusOK)
		}
		muS.Unlock()
	} else {
		rw.WriteHeader(http.StatusUnauthorized)
	}
}

func logoutHandler(rw http.ResponseWriter, req *http.Request) {
	cookieValueAndSessionKey, ok := req.Context().Value(ctxKey).(string)
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	_, ok = sessions[cookieValueAndSessionKey]
	if !ok {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	delete(sessions, cookieValueAndSessionKey)
	cookie := genCookie("", true)
	http.SetCookie(rw, cookie)
}

func entitiesHandler(rw http.ResponseWriter, req *http.Request) {
	type Entity struct {
		Id    string
		Value string
	}
	type Response struct {
		Results []Entity
		Count   uint
	}
	byteArr, err := json.Marshal(Response{Count: 15, Results: []Entity{{Id: "AAAAA", Value: "AAAAA"}, {Id: "BBBBB", Value: "BBBBB"}, {Id: "CCCCC", Value: "CCCCC"}}})
	if err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		return
	}
	rw.Write(byteArr)
}

// main

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", middleware(http.NotFound, false))
	mux.HandleFunc("GET /entities", middleware(entitiesHandler, true))
	mux.HandleFunc("GET /session", middleware(sessionHandler, false))
	mux.HandleFunc("POST /login", middleware(loginHandler, false))
	mux.HandleFunc("POST /logout", middleware(logoutHandler, false))
	setCleaner(cleanerInterval)
	err := http.ListenAndServe("0.0.0.0:8080", mux)
	if err != nil {
		panic(err)
	}
}
