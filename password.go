/*
 * Copyright (c) 2016 Michael Jacobsen (github.com/mikejac)
 *
 * This file is part of ckptool.golang.
 *
 * ckptool.golang is free software: you can redistribute
 * it and/or modify it under the terms of the GNU General Public License
 * as published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * ckptool.golang is distributed in the hope that it will
 * be useful, but WITHOUT ANY WARRANTY; without even the implied warranty
 * of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with ckptool.golang.  If not,
 * see <http://www.gnu.org/licenses/>.
 *
 */

package main

import (
	"fmt"
	"strings"
	"syscall"
	"os"
	"golang.org/x/crypto/ssh/terminal"
)

//
//
func Credentials(prompt string) (password string, ok bool) {
	fmt.Fprintf(os.Stderr, prompt)
	
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Fprintln(os.Stderr)
	
	if err != nil {
		return "", false
	}
	
	password = string(bytePassword)
	
	return strings.TrimSpace(password), true
}
