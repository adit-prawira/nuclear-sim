package reactor

import "math"

/*
Controlling Rod insertion, where the insertion percentage feeds into k-eff, causing:
	- Power inscrease when withdrawn
	- Powert decrease when inserted

Physics rule:
	- Every rod insertion reduces k-eff by a small worth value
	- Total of 211 rod inserted worth of -0.030 Δk
	- Graphite as moderator on the tip of the rod, leading to positive k-eff.
		This happens when all 211 control rods are fullwithdrawn, and every insertion of control rod will cause +0.003 into k-eff value.
		This activity will happen briefly after the rod is inserted, before the absorber takes an affect to cause Xenon poisoning
	- Minium of 15 rods must be inserted into the core, if less than that, then a WARNING must be triggered.
*/
const (
	TotalRods = 211 // Total nominal rods used for RBMK model
	MinimumSafeRods = 15 // minium rods to be inserted into the core 
	TotalRodWorth = 0.030 // Total of 0.030  Δk  when all 211 rods are inserted
	GraphiteTipSpike = 0.003 // Every rod insertion adds +0.003 into k-eff value 
)

func (r *reactor) TotalInsertedRods() int {
	return r.controlRod.RodsInserted
}

/*
Inserting control rods into the core, returning 
	- Total control rods inserted 
	- Status if the insertion of the rods causing graphite tip spike. 
		- This means when all rods are fully withdrawn, neutron atoms are experiencing
			fission that is aggressive causing power spike, and inserting graphite tip rods 
			will cause another brief spike to the thermal power.
*/
func (r *reactor) InsertRods(count int) (inserted int, isGraphiteTipSpike bool) {

	currentInsertedRods := r.controlRod.RodsInserted
	newInsertedRods := currentInsertedRods + count
	
	// Upper bound of total rods that exist in RBMK model 
	newInsertedRods = int(math.Min(float64(newInsertedRods), float64(TotalRods)))	
	r.controlRod.RodsInserted = newInsertedRods 
	
	deltaRods := newInsertedRods - currentInsertedRods
	isSpike := count > 0 && currentInsertedRods==0 

	return deltaRods, isSpike
}

/*
Withdrawing control rods from the core, returning 
	- Total rods withdrawn 
	- Status if roads being withdrawn makes total rods inserted in the core to be below 15 which 
		is the minimum safe rods 
*/
func (r *reactor) WithdrawnRods(count int) (withdrawn int, isBelowSafe bool) {
	currentInsertedRods := r.controlRod.RodsInserted
	newInsertedRods := currentInsertedRods - count
	
	// Lower bound of total rods could not be below 0
	newInsertedRods  = int(math.Min(float64(newInsertedRods), float64(0)))
	r.controlRod.RodsInserted = newInsertedRods

	deltaRods := currentInsertedRods - newInsertedRods
	isSafe := newInsertedRods < MinimumSafeRods
	return deltaRods, isSafe
}

/*
When N-rods are inserted, the aborder will cause xenon poisoning which will 
prevent neutrons atom to collides and prevent fission.
*/
func (r *reactor) RodReactivity() float64 {
	return  -float64(r.controlRod.RodsInserted) * r.rodWorthPerRod() 
}

func (r *reactor) rodWorthPerRod() float64 {
	return TotalRodWorth / float64(TotalRods)
} 



