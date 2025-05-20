# Forest Fire Simulation

This project implements a forest fire simulation using the Go programming language and the SDL2 library for graphical visualization. The simulation models the spread of fire through a randomly generated forest, taking into account factors such as wind direction, tree density, and burning stages. The main goal is to provide a way to observe how fire propagates in a forest environment under different conditions.

## Design and Implementation

The simulation represents the forest as a two-dimensional grid, where each cell corresponds to a tile that can be empty, contain a tree, be on fire, or represent various burning stages. The grid is visualized using SDL2, with each tile rendered in a color corresponding to its state.

### Forest Generation

At the start of the simulation, the forest is generated randomly based on a configurable tree density. The placement of trees is randomized to create a natural-looking forest. The user can regenerate the forest at any time using a keyboard shortcut.

### Fire Spread Mechanics

The fire spreads from burning trees to adjacent trees in each simulation step. The spread is influenced by wind, which can be configured to blow in a specific direction or randomly. The probability of fire spreading in the direction of the wind is higher, simulating real-world conditions. Each burning tree progresses through several burning stages before becoming a burned tree.

### Multithreading

To improve performance, especially for large forests, the simulation update loop is parallelized using goroutines. The grid is divided into chunks, and each chunk is processed concurrently. Though the bottleneck is primarily in the rendering phase which cannot be parallelized due to SDL2 limitations.

### User Interaction

The ways to interact with the simulation include:
- Pressing `Q` quits the simulation.
- Pressing `R` regenerates the forest randomly.
- Pressing `T` triggers a thunderbolt, igniting a random tree (optionally limited to the center of the forest for clarity).
- Pressing `W` toggles wind direction between random angle and radial.

## Goals and Applications

This project serves as an example of using Go for real-time simulations and was also a great way to learn a little of SDL2.