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

package ethereum

// SignatureLength is the length of the signature returned by the Signature
// function from the Signer interface.
const SignatureLength = 65

type Signer interface {
	// Address returns account's address used to sign data. May be empty if
	// the signer is used only to verify signatures.
	Address() Address
	// SignTransaction signs transaction. Signed transaction will be set
	// to the SignedTx field in the Transaction structure.
	SignTransaction(transaction *Transaction) error
	// Signature signs the hash of the given data and returns it.
	Signature(data []byte) ([]byte, error)
	// Recover returns the wallet address that created the given signature.
	Recover(signature []byte, data []byte) (*Address, error)
}
