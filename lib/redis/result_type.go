package redis

import "strconv"

/* * * * * * * * * * * * */
/* * Redis Result Type * */
/* * * * * * * * * * * * */

type ClusterSlotResult []interface{}

func (c *ClusterSlotResult) LimitTo(nbOfSlot int) (slotRange [][]string) {
	remainingSlot := nbOfSlot

	for _, res := range *c {
		nestedRes := res.([]interface{})
		if len(nestedRes) >= 2 {
			start := int(nestedRes[0].(int64))
			end := int(nestedRes[1].(int64))

			if remainingSlot == 0 {
				return
			}

			diff := end - start + 1
			if diff > remainingSlot {
				end = (start + remainingSlot - 1)
				slotRange = append(slotRange, []string{strconv.Itoa(start), strconv.Itoa(end)})
				return
			} else {
				slotRange = append(slotRange, []string{strconv.Itoa(start), strconv.Itoa(end)})
				remainingSlot -= diff
			}
		}
	}

	return
}
