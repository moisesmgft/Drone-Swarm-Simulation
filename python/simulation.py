import random

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
        start_i = random.randint(max(0, start[0] - 7), min(n-1, start[0] + 7))  
        start_j = random.randint(max(0, start[1] - 7), min(n-1, start[1] + 7))  
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
        atk_time = random.randint(start, end)
        if attacked in s or atk_time in mp:
            continue
        s.add(attacked)
        mp[atk_time] = attacked
    return mp

if __name__ == '__main__':
    SIZE = 50
    DRONES = 5
    ATTACKS = 2

    start, end = get_start_and_end(SIZE)
    path, moves = generate_path(start, end)  # This is the mission
    drones = generate_random(DRONES, start, SIZE)
    drones = [start] + drones  # Here, we maintain all drone positions
    attacks = generate_attacks(ATTACKS, len(moves), DRONES)

    attacked = [False] * (DRONES + 1)
    print(start)
    print(end)
    print(attacks)
    
    for t, (di, dj) in enumerate(moves):
        print(t, end=": ")
        if t in attacks:
            attacked[attacks[t]] = True
        for i in range(DRONES + 1):
            if attacked[i]:
                drones[i][1] = max(drones[i][1] - 1, 0)
            else:
                drones[i][0] = max(drones[i][0] + di, 0)
                drones[i][1] = min(drones[i][1] + dj, SIZE - 1)
            print(drones[i], end=" ")
        print()
