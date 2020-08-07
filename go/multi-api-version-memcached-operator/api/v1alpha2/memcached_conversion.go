/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

/*
For imports, we'll need the controller-runtime
[`conversion`](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/conversion)
package, plus the API version for our hub type (v1), and finally some of the
standard packages.
*/
import (
	"fmt"
	"strconv"
	"strings"

	cachev1alpha1 "github.com/example-inc/memcached-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

/*
Our "spoke" versions need to implement the
[`Convertible`](https://godoc.org/sigs.k8s.io/controller-runtime/pkg/conversion#Convertible)
interface.  Namely, they'll need `ConvertTo` and `ConvertFrom` methods to convert to/from
the hub version.
*/

/*
ConvertTo is expected to modify its argument to contain the converted object.
Most of the conversion is straightforward copying, except for converting our changed field.
*/
// ConvertTo converts this Memcached to the Hub version (v1alpha1).
func (src *Memcached) ConvertTo(dstRaw conversion.Hub) error {
	switch t := dstRaw.(type) {
	case *cachev1alpha1.Memcached:
		dst := dstRaw.(*cachev1alpha1.Memcached)

		// conversion implementation goes here
		// in our case, we convert the price in structured form to string form.
		// Spec
		dst.Spec.Price = fmt.Sprintf("%d %s", src.Spec.Price.Amount, src.Spec.Price.Currency)
		sched := src.Spec.Schedule
		scheduleParts := []string{"*", "*", "*", "*", "*"}
		if sched.Minute != nil {
			scheduleParts[0] = string(*sched.Minute)
		}
		if sched.Hour != nil {
			scheduleParts[1] = string(*sched.Hour)
		}
		if sched.DayOfMonth != nil {
			scheduleParts[2] = string(*sched.DayOfMonth)
		}
		if sched.Month != nil {
			scheduleParts[3] = string(*sched.Month)
		}
		if sched.DayOfWeek != nil {
			scheduleParts[4] = string(*sched.DayOfWeek)
		}
		dst.Spec.Schedule = strings.Join(scheduleParts, " ")
		// ObjectMeta
		dst.ObjectMeta = src.ObjectMeta
		// rest of conversion
		dst.Spec.Size = src.Spec.Size
		dst.Status.Nodes = src.Status.Nodes

		return nil
	default:
		return fmt.Errorf("unsupported type %v", t)
	}
}

/*
ConvertFrom is expected to modify its receiver to contain the converted object.
Most of the conversion is straightforward copying, except for converting our changed field.
*/
// ConvertFrom converts from the Hub version (v1alpha1) to this version.
func (dst *Memcached) ConvertFrom(srcRaw conversion.Hub) error {
	switch t := srcRaw.(type) {
	case *cachev1alpha1.Memcached:
		src := srcRaw.(*cachev1alpha1.Memcached)

		// conversion implementation goes here
		// We parse price amount and currency from the string form and
		// convert it in structured form.
		parts := strings.Fields(src.Spec.Price)
		if len(parts) != 2 {
			return fmt.Errorf("invalid price")
		}
		amount, err := strconv.Atoi(parts[0])
		if err != nil {
			return err
		}
		dst.Spec.Price = Price{
			Amount:   int64(amount),
			Currency: parts[1],
		}
		schedParts := strings.Split(src.Spec.Schedule, " ")
		if len(schedParts) != 5 {
			return fmt.Errorf("invalid schedule: not a standard 5-field schedule")
		}
		partIfNeeded := func(raw string) *CronField {
			if raw == "*" {
				return nil
			}
			part := CronField(raw)
			return &part
		}
		dst.Spec.Schedule.Minute = partIfNeeded(schedParts[0])
		dst.Spec.Schedule.Hour = partIfNeeded(schedParts[1])
		dst.Spec.Schedule.DayOfMonth = partIfNeeded(schedParts[2])
		dst.Spec.Schedule.Month = partIfNeeded(schedParts[3])
		dst.Spec.Schedule.DayOfWeek = partIfNeeded(schedParts[4])
		// ObjectMeta
		dst.ObjectMeta = src.ObjectMeta
		//rest of the conversion
		dst.Spec.Size = src.Spec.Size
		dst.Status.Nodes = src.Status.Nodes

		return nil
	default:
		return fmt.Errorf("unsupported type %v", t)
	}
}
