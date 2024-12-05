import pygame
import sys
import select

# Constants
WIDTH, HEIGHT = 2000, 2000
GRID_SIZE = 50
DRONE_RADIUS = 10
BASE_SIZE = 20
FPS = 60
WHITE = (255, 255, 255)
BLACK = (0, 0, 0)
BLUE = (0, 0, 255)
GREEN = (0, 255, 0)
RED = (255, 0, 0)

# Initialize pygame
pygame.init()
screen = pygame.display.set_mode((WIDTH, HEIGHT))
pygame.display.set_caption("Drones and Central Base")
clock = pygame.time.Clock()

# Drone and base information
drones = []  # List of drones, each entry is [x, y]
connections = []  # [(id1, id2), ...]
base_connections = []  # [id1, id2, ...]

def draw_grid():
    """Draw the grid on the screen."""
    for x in range(0, WIDTH, GRID_SIZE):
        pygame.draw.line(screen, BLACK, (x, 0), (x, HEIGHT))
    for y in range(0, HEIGHT, GRID_SIZE):
        pygame.draw.line(screen, BLACK, (0, y), (WIDTH, y))

def draw_drones_and_base():
    """Draw the drones, base, and their connections."""
    # Draw connections between drones
    for id1, id2 in connections:
        if id1 < len(drones) and id2 < len(drones):
            x1, y1 = drones[id1]
            x2, y2 = drones[id2]
            pygame.draw.line(screen, RED, (x1, y1), (x2, y2), 2)
    
    # Draw connections to the base
    for drone_id in base_connections:
        if drone_id < len(drones):
            x, y = drones[drone_id]
            pygame.draw.line(screen, GREEN, (WIDTH // 2, HEIGHT // 2), (x, y), 2)

    # Draw drones
    for id, (x, y) in enumerate(drones):
        if id != 0:
            pygame.draw.circle(screen, BLUE, (x, y), DRONE_RADIUS)
        else:
            pygame.draw.circle(screen, GREEN, (x, y), DRONE_RADIUS)

def update_positions(data):
    """Update drone positions based on the input."""
    _, id, x, y = data
    drones[id] = [x * GRID_SIZE + GRID_SIZE // 2, y * GRID_SIZE + GRID_SIZE // 2]

def process_input(input_line):
    """Process the input and update the system."""
    input_line = input_line.strip()
    if input_line.startswith("POS"):
        # Handle "POS" command
        if "," in input_line:  # Handle comma-separated format
            _, id, x, y = input_line.split(", ")
        else:  # Handle space-separated format
            _, id, x, y = input_line.split()
        update_positions(("POS", int(id), int(x), int(y)))
        update_connections()

def update_connections():
    """Update drone connections dynamically based on distance."""
    global connections
    connections = []
    base_connections = []

    # Parse connections between drones
    for i in range(len(drones)):
        for j in range(i + 1, len(drones)):  # Avoid duplicate pairs
            x1, y1 = drones[i]
            x2, y2 = drones[j]
            # Calculate distance using Pythagoras
            distance = ((x2 - x1)**2 + (y2 - y1)**2)**0.5
            if distance <= 25 * GRID_SIZE:  # 5 units, scaled by GRID_SIZE
                connections.append((i, j))

    # Parse connections to the base
    base_x, base_y = WIDTH // 2, HEIGHT // 2
    for i, (x, y) in enumerate(drones):
        distance_to_base = ((x - base_x)**2 + (y - base_y)**2)**0.5
        if distance_to_base <= 25 * GRID_SIZE:  # 5 units to the base
            base_connections.append(i)

def main():
    """Main loop."""
    global drones
    drones_count = 0

    # Initial setup
    print("Enter the number of drones:")
    drones_count = int(input().strip())
    drones = [[i * GRID_SIZE + GRID_SIZE // 2, GRID_SIZE // 2] for i in range(drones_count)]

    running = True
    while running:
        for event in pygame.event.get():
            if event.type == pygame.QUIT:
                running = False

        # Check for new input
        if sys.stdin in select.select([sys.stdin], [], [], 0)[0]:
            input_line = sys.stdin.readline()
            if input_line:
                process_input(input_line)

        # Draw everything
        screen.fill(WHITE)
        draw_grid()
        draw_drones_and_base()
        pygame.display.flip()
        clock.tick(FPS)

    pygame.quit()

if __name__ == "__main__":
    main()
