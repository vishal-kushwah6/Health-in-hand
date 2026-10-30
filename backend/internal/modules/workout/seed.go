package workout

import (
	"context"
	"time"
)

// SeedExercises fills the built-in library on first run. It does nothing if
// built-in exercises already exist, so it is safe to call on every startup.
func SeedExercises(ctx context.Context, repo ExerciseRepository) error {
	n, err := repo.CountGlobal(ctx)
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	list := seedList()
	now := time.Now().UTC()
	for i := range list {
		list[i].CreatedAt = now
		list[i].Archived = false
	}
	return repo.InsertMany(ctx, list)
}

func mg(m ...MuscleGroup) []MuscleGroup { return m }

func ex(name string, eq Equipment, d Difficulty, muscles []MuscleGroup, desc string, steps ...string) Exercise {
	return Exercise{Name: name, Equipment: eq, Difficulty: d, MuscleGroups: muscles, Description: desc, Instructions: steps}
}

func seedList() []Exercise {
	return []Exercise{
		// ---- Chest ----
		ex("Bench Press", EquipBarbell, DifficultyIntermediate, mg(MuscleChest, MuscleTriceps, MuscleShoulders),
			"The classic barbell press for chest strength.",
			"Lie on the bench with eyes under the bar and feet flat.", "Lower the bar to mid-chest, elbows about 45 degrees.", "Press up until your arms are straight."),
		ex("Incline Dumbbell Press", EquipDumbbell, DifficultyIntermediate, mg(MuscleChest, MuscleShoulders),
			"Targets the upper chest.",
			"Set the bench to 30 degrees and hold a dumbbell in each hand.", "Press the weights up and slightly together, then lower under control."),
		ex("Push Up", EquipBodyweight, DifficultyBeginner, mg(MuscleChest, MuscleTriceps, MuscleCore),
			"Bodyweight chest and triceps staple.",
			"Hands slightly wider than shoulders, body in a straight line.", "Lower your chest to the floor, then push back up."),
		ex("Dumbbell Fly", EquipDumbbell, DifficultyIntermediate, mg(MuscleChest),
			"Isolation move that stretches the chest.",
			"Lie flat with dumbbells above your chest, elbows soft.", "Open your arms in a wide arc, then bring them back together."),
		ex("Cable Crossover", EquipCable, DifficultyIntermediate, mg(MuscleChest),
			"Constant tension chest finisher.",
			"Stand between high pulleys with a slight forward lean.", "Sweep the handles down and together in front of you."),
		ex("Machine Chest Press", EquipMachine, DifficultyBeginner, mg(MuscleChest, MuscleTriceps),
			"A guided press that is easy to learn.",
			"Adjust the seat so handles are at mid-chest.", "Press forward until arms are nearly straight, then return slowly."),

		// ---- Back ----
		ex("Pull Up", EquipBodyweight, DifficultyIntermediate, mg(MuscleBack, MuscleBiceps),
			"Vertical pull that builds a wide back.",
			"Hang from the bar with an overhand grip a little wider than shoulders.", "Pull your chest toward the bar, then lower fully."),
		ex("Lat Pulldown", EquipCable, DifficultyBeginner, mg(MuscleBack, MuscleBiceps),
			"A pull-up alternative with adjustable weight.",
			"Grip the bar wide and sit with thighs secured.", "Pull the bar to your upper chest, then return slowly."),
		ex("Barbell Row", EquipBarbell, DifficultyIntermediate, mg(MuscleBack, MuscleBiceps),
			"Heavy horizontal pull for back thickness.",
			"Hinge at the hips with a flat back, bar hanging at arm's length.", "Row the bar to your lower ribs, then lower it."),
		ex("Seated Cable Row", EquipCable, DifficultyBeginner, mg(MuscleBack, MuscleBiceps),
			"Controlled horizontal pull.",
			"Sit tall with a slight knee bend and grip the handle.", "Pull to your stomach, squeeze your shoulder blades, return slowly."),
		ex("One-Arm Dumbbell Row", EquipDumbbell, DifficultyBeginner, mg(MuscleBack, MuscleBiceps),
			"Unilateral row to fix side-to-side imbalances.",
			"Support yourself on a bench with one hand and knee.", "Row the dumbbell to your hip, then lower it."),
		ex("Deadlift", EquipBarbell, DifficultyAdvanced, mg(MuscleBack, MuscleHamstring, MuscleGlutes),
			"Full-body pull from the floor.",
			"Stand with the bar over mid-foot and grip just outside your legs.", "Brace your core and push the floor away, standing tall.", "Lower the bar along your legs with a flat back."),
		ex("Face Pull", EquipCable, DifficultyBeginner, mg(MuscleShoulders, MuscleBack),
			"Rear shoulder and upper back health move.",
			"Set the rope at face height.", "Pull toward your face, elbows high, then return slowly."),

		// ---- Shoulders ----
		ex("Overhead Press", EquipBarbell, DifficultyIntermediate, mg(MuscleShoulders, MuscleTriceps),
			"Standing barbell press for strong shoulders.",
			"Hold the bar at collarbone height with a tight core.", "Press overhead until arms lock out, then lower under control."),
		ex("Dumbbell Shoulder Press", EquipDumbbell, DifficultyBeginner, mg(MuscleShoulders, MuscleTriceps),
			"Seated or standing dumbbell press.",
			"Start with dumbbells at shoulder height.", "Press up and slightly inward, then lower slowly."),
		ex("Lateral Raise", EquipDumbbell, DifficultyBeginner, mg(MuscleShoulders),
			"Builds the side of the shoulders.",
			"Hold light dumbbells at your sides.", "Raise your arms out to shoulder height, then lower slowly."),
		ex("Rear Delt Fly", EquipDumbbell, DifficultyBeginner, mg(MuscleShoulders, MuscleBack),
			"Targets the back of the shoulders.",
			"Hinge forward with a flat back, dumbbells hanging.", "Raise your arms out to the sides, then lower."),
		ex("Pike Push Up", EquipBodyweight, DifficultyIntermediate, mg(MuscleShoulders, MuscleTriceps),
			"Bodyweight shoulder press variation.",
			"Start in a push-up position with hips high.", "Lower your head toward the floor, then press back up."),

		// ---- Biceps ----
		ex("Barbell Curl", EquipBarbell, DifficultyBeginner, mg(MuscleBiceps),
			"Standard mass builder for the biceps.",
			"Stand tall holding the bar with an underhand grip.", "Curl to your shoulders without swinging, then lower."),
		ex("Dumbbell Curl", EquipDumbbell, DifficultyBeginner, mg(MuscleBiceps),
			"Alternating or simultaneous curls.",
			"Keep elbows at your sides.", "Curl the weights up, squeeze, and lower slowly."),
		ex("Hammer Curl", EquipDumbbell, DifficultyBeginner, mg(MuscleBiceps, MuscleForearms),
			"Neutral grip curl for biceps and forearms.",
			"Hold dumbbells with palms facing each other.", "Curl up and lower with control."),
		ex("Cable Curl", EquipCable, DifficultyBeginner, mg(MuscleBiceps),
			"Curl with constant tension.",
			"Stand facing a low pulley with a straight or EZ bar.", "Curl up and lower slowly."),

		// ---- Triceps ----
		ex("Tricep Pushdown", EquipCable, DifficultyBeginner, mg(MuscleTriceps),
			"Simple isolation for the triceps.",
			"Stand at a high pulley, elbows pinned to your sides.", "Push the bar down until arms are straight, then return."),
		ex("Skull Crusher", EquipBarbell, DifficultyIntermediate, mg(MuscleTriceps),
			"Lying triceps extension.",
			"Lie on a bench holding the bar above your chest.", "Bend your elbows to lower the bar toward your forehead, then extend."),
		ex("Overhead Tricep Extension", EquipDumbbell, DifficultyBeginner, mg(MuscleTriceps),
			"Stretches the long head of the triceps.",
			"Hold one dumbbell overhead with both hands.", "Lower it behind your head, then extend your arms."),
		ex("Close Grip Bench Press", EquipBarbell, DifficultyIntermediate, mg(MuscleTriceps, MuscleChest),
			"Bench press variation that emphasises triceps.",
			"Grip the bar about shoulder width.", "Lower to your lower chest with elbows tucked, then press."),
		ex("Bench Dip", EquipBodyweight, DifficultyBeginner, mg(MuscleTriceps, MuscleShoulders),
			"Bodyweight triceps exercise using a bench.",
			"Place your hands on a bench behind you, legs forward.", "Lower your hips by bending your elbows, then press back up."),

		// ---- Legs ----
		ex("Back Squat", EquipBarbell, DifficultyIntermediate, mg(MuscleQuads, MuscleGlutes, MuscleHamstring),
			"The foundation lower-body strength lift.",
			"Rest the bar on your upper back, feet shoulder width.", "Sit down and back until thighs are parallel or lower.", "Drive through your feet to stand."),
		ex("Goblet Squat", EquipDumbbell, DifficultyBeginner, mg(MuscleQuads, MuscleGlutes),
			"Beginner-friendly squat holding a weight at the chest.",
			"Hold a dumbbell against your chest.", "Squat down between your knees, then stand."),
		ex("Leg Press", EquipMachine, DifficultyBeginner, mg(MuscleQuads, MuscleGlutes),
			"Machine press for leg strength.",
			"Place feet shoulder width on the platform.", "Lower until knees reach about 90 degrees, then press without locking out."),
		ex("Walking Lunge", EquipDumbbell, DifficultyBeginner, mg(MuscleQuads, MuscleGlutes),
			"Dynamic single-leg work.",
			"Step forward and lower your back knee toward the floor.", "Push through the front foot and step into the next rep."),
		ex("Bulgarian Split Squat", EquipDumbbell, DifficultyAdvanced, mg(MuscleQuads, MuscleGlutes),
			"Rear-foot-elevated single-leg squat.",
			"Rest your back foot on a bench behind you.", "Lower straight down, then drive up through the front foot."),
		ex("Romanian Deadlift", EquipBarbell, DifficultyIntermediate, mg(MuscleHamstring, MuscleGlutes),
			"Hip hinge that builds the hamstrings.",
			"Hold the bar at hip height with soft knees.", "Push your hips back, lowering the bar along your legs, then stand."),
		ex("Leg Extension", EquipMachine, DifficultyBeginner, mg(MuscleQuads),
			"Quad isolation.",
			"Adjust the pad above your ankles.", "Extend your legs fully, then lower slowly."),
		ex("Leg Curl", EquipMachine, DifficultyBeginner, mg(MuscleHamstring),
			"Hamstring isolation.",
			"Position the pad above your heels.", "Curl your heels toward your glutes, then return slowly."),
		ex("Hip Thrust", EquipBarbell, DifficultyIntermediate, mg(MuscleGlutes, MuscleHamstring),
			"Top glute-building exercise.",
			"Rest your upper back on a bench with the bar over your hips.", "Drive your hips up until your body is straight, then lower."),
		ex("Standing Calf Raise", EquipMachine, DifficultyBeginner, mg(MuscleCalves),
			"Builds the calves.",
			"Place the balls of your feet on the platform.", "Rise onto your toes, pause, then lower for a full stretch."),
		ex("Glute Bridge", EquipBodyweight, DifficultyBeginner, mg(MuscleGlutes, MuscleHamstring),
			"Floor-based glute activation.",
			"Lie on your back with knees bent.", "Lift your hips until your body forms a line, squeeze, then lower."),

		// ---- Core ----
		ex("Plank", EquipBodyweight, DifficultyBeginner, mg(MuscleCore),
			"Isometric core hold.",
			"Support yourself on forearms and toes with a straight body.", "Hold without letting your hips sag."),
		ex("Crunch", EquipBodyweight, DifficultyBeginner, mg(MuscleCore),
			"Basic abdominal flexion.",
			"Lie on your back with knees bent.", "Curl your shoulders off the floor, then lower."),
		ex("Hanging Leg Raise", EquipBodyweight, DifficultyAdvanced, mg(MuscleCore),
			"Advanced lower-abs exercise.",
			"Hang from a bar with straight arms.", "Raise your legs to hip height or higher, then lower slowly."),
		ex("Russian Twist", EquipBodyweight, DifficultyBeginner, mg(MuscleCore),
			"Rotational core work.",
			"Sit leaning back with feet slightly raised.", "Rotate your torso side to side."),
		ex("Ab Wheel Rollout", EquipNone, DifficultyAdvanced, mg(MuscleCore, MuscleShoulders),
			"Demanding anti-extension core exercise.",
			"Kneel holding the wheel under your shoulders.", "Roll forward as far as you can with a tight core, then pull back."),
		ex("Mountain Climber", EquipBodyweight, DifficultyBeginner, mg(MuscleCore, MuscleCardio),
			"Core and cardio in one.",
			"Start in a high plank.", "Drive your knees toward your chest alternately at a steady pace."),

		// ---- Full body / cardio ----
		ex("Burpee", EquipBodyweight, DifficultyIntermediate, mg(MuscleFullBody, MuscleCardio),
			"High-intensity full-body movement.",
			"Drop into a push-up, then jump your feet back in.", "Explode upward into a jump."),
		ex("Kettlebell Swing", EquipKettlebell, DifficultyIntermediate, mg(MuscleFullBody, MuscleGlutes, MuscleHamstring),
			"Explosive hip hinge.",
			"Hinge back and swing the kettlebell between your legs.", "Snap your hips forward to float it to chest height."),
		ex("Running", EquipNone, DifficultyBeginner, mg(MuscleCardio, MuscleQuads, MuscleCalves),
			"Steady-state or interval running.",
			"Warm up with an easy jog.", "Keep a pace you can hold, and cool down by walking."),
		ex("Jump Rope", EquipNone, DifficultyBeginner, mg(MuscleCardio, MuscleCalves),
			"Efficient cardio and coordination work.",
			"Keep elbows close and turn the rope with your wrists.", "Land softly on the balls of your feet."),
		ex("Rowing Machine", EquipMachine, DifficultyBeginner, mg(MuscleCardio, MuscleBack),
			"Low-impact full-body cardio.",
			"Push with your legs first, then lean back and pull.", "Return in reverse order: arms, body, legs."),
		ex("Farmer's Carry", EquipDumbbell, DifficultyBeginner, mg(MuscleForearms, MuscleCore, MuscleFullBody),
			"Loaded carry for grip and trunk stability.",
			"Hold a heavy dumbbell in each hand at your sides.", "Walk tall with short, controlled steps."),
	}
}
