// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package config

// Option is implemented by resolved configuration value types that can
// convert themselves back into a functional option of type T.
//
// This allows a value that has already been applied to one component to be
// forwarded to another without unpacking its underlying fields. For example,
// [BlogRoot] and [Logger] both implement Option[BaseOption], so a server can
// forward its resolved BlogRoot or Logger to a sub-component by calling
// AsOption() rather than constructing a new option from scratch.
type Option[T any] interface {
	AsOption() T
}
