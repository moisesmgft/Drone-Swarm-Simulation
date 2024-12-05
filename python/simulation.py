import random
import time
import pygame
import sys
import select

# Constants
WIDTH, HEIGHT = 2000, 2000
GRID_SIZE = 4
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
endx = 0
endy = 0


# Drone and base information
drones = []  # List of drones, each entry is [x, y]
connections = []  # [(id1, id2), ...]
disconnects = []
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
    pygame.draw.circle(screen, (255,165,0), (endx*GRID_SIZE, endy*GRID_SIZE), DRONE_RADIUS*30)
    for id1, id2 in connections:
        if id1 < len(drones) and id2 < len(drones):
            x1, y1 = drones[id1]
            x2, y2 = drones[id2]
            pygame.draw.line(screen, RED, (x1*GRID_SIZE, y1*GRID_SIZE), (x2*GRID_SIZE, y2*GRID_SIZE), 2)
    
    # Draw connections to the base
    for drone_id in base_connections:
        if drone_id < len(drones):
            x, y = drones[drone_id]
            pygame.draw.line(screen, GREEN, (WIDTH // 2, HEIGHT // 2), (x*GRID_SIZE, y*GRID_SIZE), 2)

    # Draw drones
    for id, (x, y) in enumerate(drones):
        if id != 0:
            pygame.draw.circle(screen, BLUE, (x*GRID_SIZE, y*GRID_SIZE), DRONE_RADIUS)
        else:
            pygame.draw.circle(screen, GREEN, (x*GRID_SIZE, y*GRID_SIZE), DRONE_RADIUS)
    
    

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
            if distance <= 25*GRID_SIZE and not attacked[i] and not attacked[j]:  # 5 units, scaled by GRID_SIZE
                connections.append((i, j))



def get_start_and_end(n):
    # Start position
    start_i = random.randint(int(0.8 * n), n - 1)  
    start_j = random.randint(0, int(0.2 * n)) 
    start = [start_i, start_j]

    # End position
    end_i = random.randint(0, int(0.2 * n))   
    end_j = random.randint(int(0.8 * n), n - 1)  
    end = [end_i, end_j]
    
    return start, end

def generate_random(drones, start, n):
    ret = []
    for _ in range(drones):
        # Start position
        start_i = random.randint(max(0, start[0] - 7*GRID_SIZE), min(n-1, start[0] + 7*GRID_SIZE))  
        start_j = random.randint(max(0, start[1] - 7*GRID_SIZE), min(n-1, start[1] + 7*GRID_SIZE))  
        ret.append([start_i, start_j])
    return ret

def generate_path(start, end):
    cur = start.copy()
    path = [cur[:]]
    moves = [[0, 0]]
    while cur != end:
        mv = []
        if cur[0] != end[0]:
            mv.append([-1 if cur[0] > end[0] else 1, 0])
        if cur[1] != end[1]:
            mv.append([0, -1 if cur[1] > end[1] else 1])

        move = random.choice(mv)
        cur[0] += move[0]
        cur[1] += move[1]

        path.append(cur[:])
        moves.append(move[:])
    
    return path, moves

def generate_attacks(n, time, total):
    start = 10
    end = time - 20
    assert start < end
    s = set()
    mp = {}
    while len(mp) < n:
        attacked = random.randint(1, total)
        atk_time = random.randint(start, int(end*0.75))
        if attacked in s or atk_time in mp:
            continue
        s.add(attacked)
        mp[atk_time] = attacked
    return mp

if __name__ == '__main__':
    SIZE = 500
    DRONES = 5
    ATTACKS = random.randint(0, DRONES-1)

    start, end = get_start_and_end(SIZE)
    (endx, endy) = end
    path, moves = generate_path(start, end)  # This is the mission
    drones = generate_random(DRONES, start, SIZE)
    drones = [start] + drones  # Here, we maintain all drone positions
    attacks = generate_attacks(ATTACKS, len(moves), DRONES)

    attacked = [False] * (DRONES + 1)
    print(start)
    print(end)
    print(attacks)
    draw_drones_and_base()
    for t, (di, dj) in enumerate(moves):
        if t in attacks:
            attacked[attacks[t]] = True

        for i in range(DRONES + 1):
            if not attacked[i]:
                multrand = 1/3
                isNZ = i != 0
                drones[i][0] = max(drones[i][0] + di + isNZ*random.randint(-int(GRID_SIZE*multrand), int(GRID_SIZE*multrand)), 0)
                drones[i][1] = min(drones[i][1] + dj + isNZ*random.randint(-int(GRID_SIZE*multrand), int(GRID_SIZE*multrand)), SIZE - 1)
                update_connections()
            # print(drones[i], end=" ")
        screen.fill(WHITE)
        # draw_grid()
        draw_drones_and_base()
        pygame.display.flip()
        clock.tick(FPS)
        time.sleep(0.01)

    if ATTACKS >= 3:
        screen.fill(RED)
        #WRITE SUCCESS VERY BIG IN THE MIDDLE   
        font = pygame.font.Font('freesansbold.ttf', 50)
        text = font.render('FAIL', True, (0,0,0), RED)
        textRect = text.get_rect()
        textRect.center = (WIDTH // 2, HEIGHT // 2)
        screen.blit(text, textRect)
        pygame.display.flip()
        time.sleep(5)
    else:
        screen.fill(GREEN)
        #WRITE SUCCESS VERY BIG IN THE MIDDLE
        font = pygame.font.Font('freesansbold.ttf', 50)
        text = font.render('SUCCESS', True, (0,0,0), GREEN)
        textRect = text.get_rect()
        textRect.center = (WIDTH // 2, HEIGHT // 2)
        screen.blit(text, textRect)
        pygame.display.flip()
        time.sleep(5)

        
        print()
