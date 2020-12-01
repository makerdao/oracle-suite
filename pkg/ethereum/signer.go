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

// SignatureLength is the expected length of the Signature.
const SignatureLength = 65

// Signature represents the 65 byte signature.
type Signature [SignatureLength]byte

func SignatureFromBytes(b []byte) Signature {
	var s Signature
	copy(s[:], b)
	return s
}

func SignatureFromVRS(v uint8, r [32]byte, s [32]byte) Signature {
	return SignatureFromBytes(append(append(append([]byte{}, r[:]...), s[:]...), v))
}

func (s Signature) VRS() (sv uint8, sr [32]byte, ss [32]byte) {
	copy(sr[:], s[:32])
	copy(ss[:], s[32:64])
	sv = s[64]
	return
}

func (s Signature) Bytes() []byte {
	return s[:]
}

type Signer interface {
	// Address returns account's address used to sign data. May be empty if
	// the signer is used only to verify signatures.
	Address() Address
	// SignTransaction signs transaction. Signed transaction will be set
	// to the SignedTx field in the Transaction structure.
	SignTransaction(transaction *Transaction) error
	// Signature signs the hash of the given data and returns it.
	Signature(data []byte) (Signature, error)
	// Recover returns the wallet address that created the given signature.
	Recover(signature Signature, data []byte) (*Address, error)
}
