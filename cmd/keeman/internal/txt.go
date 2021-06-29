//  Copyright (C) 2020 Maker Ecosystem Growth Holdings, INC.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package internal

import (
	"bufio"
	"log"
	"os"
	"strings"
	"unicode"
)

func ReadLineOrSame(filename string) string {
	lines, err := readLines(filename, 1)
	if err != nil {
		return filename
	}
	if len(lines) == 0 {
		return ""
	}
	return lines[0]
}

func readLines(filename string, limit int) ([]string, error) {
	file, closeFile, err := openFile(filename)
	if err != nil {
		return nil, err
	}
	defer closeFile()
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		s := strings.Trim(stripCommentString(scanner.Text()), "\t \n")
		if len(s) != 0 {
			lines = append(lines, s)
		}
		if 0 < limit && limit <= len(lines) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}
func openFile(filename string) (*os.File, func(), error) {
	file, err := os.Open(filename)
	return file, func() {
		if err := file.Close(); err != nil {
			log.Fatal(file.Name(), err)
		}
	}, err
}
func stripCommentString(s string) string {
	if cut := strings.IndexAny(s, "#;"); cut >= 0 {
		return strings.TrimRightFunc(s[:cut], unicode.IsSpace)
	}
	return s
}
