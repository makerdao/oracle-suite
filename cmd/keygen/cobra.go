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

package main

import (
	"fmt"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "keygen",
		Version:       "DEV",
		Short:         "",
		Long:          ``,
		SilenceErrors: false,
		SilenceUsage:  true,
		RunE: func(cmd *cobra.Command, args []string) error {
			privKey, pubKey, err := crypto.GenerateKeyPair(crypto.Ed25519, 0)
			if err != nil {
				return err
			}

			pubKeyB, err := crypto.MarshalPublicKey(pubKey)
			if err != nil {
				return err
			}
			privKeyB, err := crypto.MarshalPrivateKey(privKey)
			if err != nil {
				return err
			}

			fmt.Printf("Public key:  %s\n", crypto.ConfigEncodeKey(pubKeyB))
			fmt.Printf("Private key: %s\n", crypto.ConfigEncodeKey(privKeyB))

			return nil
		},
	}

	return rootCmd
}
