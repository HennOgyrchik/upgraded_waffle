package main

import (
	"fmt"
	"net"
	"sync"
	"upgraded_waffle/postgres"
)

type user struct {
	conn  net.Conn
	login string
	ip    string
	sync.Mutex
	userList map[string]user
}

func main() {
	listener, err := net.Listen("tcp", ":4545")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer listener.Close()
	usrList := make(map[string]user)
	fmt.Println("Server is running...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			conn.Close()
			continue
		}

		go welcome(conn, usrList)
	}

}

func (u *user) read() (string, error) {
	buff := make([]byte, 1024)
	n, err := u.conn.Read(buff)
	if err != nil {
		return "", err
	}
	return string(buff[0:n]), nil
}

func (u *user) write(s string) {
	u.conn.Write([]byte(s))
}

func welcome(conn net.Conn, usrList map[string]user) {
	usr := user{conn: conn, ip: conn.RemoteAddr().String(), userList: usrList}

	for b := true; b; {
		usr.write("У вас есть учетная запись? [y/n]:")
		answ, err := usr.read()
		if err != nil {
			return
		}

		switch answ {
		case "y":
			{
				err = usr.authorization()
				if err != nil {
					return
				}

				usr.Lock()
				usr.userList[usr.login] = usr
				usr.Unlock()

				b = false

			}
		case "n":
			{
				err = usr.registration()
				if err != nil {
					return
				}
			}
		}

	}

	err := usr.listener()
	usr.Lock()
	delete(usr.userList, usr.login)
	usr.Unlock()
	if err != nil {
		return
	}
}

func (u *user) authorization() error {
	for i := 0; i < 5; i++ {
		u.write("Введите логин: ")
		login, err := u.read()
		if err != nil {
			return err
		}
		u.write("Введите пароль: ")
		pass, err := u.read()
		if err != nil {
			return err
		}

		ok, err := postgres.Authorization(login, pass)
		if err != nil {
			return err
		}
		if !ok {
			u.write("Неверный логин или пароль!\n")
			continue
		}
		_, ok = u.userList[login]
		if ok {
			u.write("Эта учетная запись уже используется!\n")
			continue
		}

		u.write("Авторизация прошла успешно!\n")
		u.login = login
		err = postgres.WriteMessage("server", u.login+" присоединился к чату!")
		if err != nil {
			return err
		}
		u.mailing()

		return err
	}
	u.write("Пока!")
	u.conn.Close()
	return nil
}

func (u *user) registration() error {
	for {

		u.write("Придумайте логин: ")
		login, err := u.read()
		if err != nil {
			return err
		}

		ok, err := postgres.CheckLogin(login)
		fmt.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!", err)
		if err != nil {
			return err
		}
		if !ok {
			u.write("Пользователь с таким логином существует\n")
			continue
		}

		u.write("Придумайте пароль: ")
		pass, err := u.read()
		if err != nil {
			return err
		}
		u.write("Повторите пароль: ")
		passVer, err := u.read()
		if err != nil {
			return err
		}
		if pass != passVer {
			u.write("Введенные пароли не совпадают\n")
			continue
		}

		err = postgres.Registration(login, pass)
		if err != nil {
			return err
		}
		u.write("Регистрация прошла успешно!\n")
		break
	}
	return nil
}

func (u user) listener() error {
	for {
		mes, err := u.read()
		if err != nil {
			return err
		}
		err = postgres.WriteMessage(u.login, mes)
		if err != nil {
			return err
		}
		u.mailing()
	}
}

func (u user) sender() {
	mes, err := postgres.GetLastMessage()
	if err != nil {
	}
	u.write(mes)
}

func (u user) mailing() {
	for i, v := range u.userList {
		if i != u.login {
			v.sender()
		}
	}

}
