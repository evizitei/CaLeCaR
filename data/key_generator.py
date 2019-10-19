import argparse
import numpy as np
import random


def generate_lru_keys(file, key_count, max_keyval):
    mean = max_keyval / 2
    generated = 0

    def sample_int_in_range(mu, minval=0, maxval=(max_keyval - 1)):
        standard_deviation = 8
        key_sample = np.random.normal(loc=mu, scale=standard_deviation)
        return int(min(maxval, max(minval, np.round(key_sample))))

    while generated < key_count:
        key = sample_int_in_range(mean)
        file.write("key{0}\n".format(key))
        mean = sample_int_in_range(mean, minval=75, maxval=(max_keyval - 25))
        generated += 1


def generate_lcr_scan_keys(file, key_count, max_keyval):
    freq_set_size = 250
    cost_set_size = 250
    all_available_keys = range(0, max_keyval)
    freq_avail_keys = range(0, 1000)
    cost_avail_keys = range(int(max_keyval - (max_keyval/10)), max_keyval)
    freq_set = set([])
    while len(freq_set) < freq_set_size:
        freq_set.add(random.choice(freq_avail_keys))
    cost_set = set([])
    while len(cost_set) < cost_set_size:
        cost_set.add(random.choice(cost_avail_keys))
    complement_set = list((set(all_available_keys) - freq_set) - cost_set)
    generated = 0
    freq_index = 0
    cost_index = 0
    freq_list = list(freq_set)
    cost_list = list(cost_set)
    while generated < (freq_set_size*3):
        selected_key = freq_list[freq_index]
        file.write("key{0}\n".format(selected_key))
        freq_index += 1
        if freq_index >= freq_set_size:
            freq_index = 0
        generated += 1
    while generated < 2 * (cost_set_size*3):
        selected_key = cost_list[cost_index]
        file.write("key{0}\n".format(selected_key))
        cost_index += 1
        if cost_index >= cost_set_size:
            cost_index = 0
        generated += 1
    scan_size = 250
    scan_step = 0
    current_phase = "SCAN"
    total_index = 0
    while generated < key_count:
        if scan_step >= scan_size:
            scan_step = 0
            if current_phase == "SCAN":
                current_phase = "FREQ"
            elif current_phase == "FREQ":
                current_phase = "COST"
            else:
                current_phase = "SCAN"
        if current_phase == "SCAN":
            selected_key = complement_set[total_index]
            file.write("key{0}\n".format(selected_key))
            total_index += 1
            if total_index >= len(complement_set):
                total_index = 0
        elif current_phase == "FREQ":
            selected_key = freq_list[freq_index]
            file.write("key{0}\n".format(selected_key))
            freq_index += 1
            if freq_index >= freq_set_size:
                freq_index = 0
        else:
            selected_key = cost_list[cost_index]
            file.write("key{0}\n".format(selected_key))
            cost_index += 1
            if cost_index >= cost_set_size:
                cost_index = 0
        generated += 1
        scan_step += 1
        print(current_phase)


def generate_lfu_scan_keys(file, key_count, max_keyval):
    available_keys = list(range(0, max_keyval))
    freq_set_size = 200
    freq_set = set([])
    while len(freq_set) < freq_set_size:
        freq_set.add(random.choice(available_keys))
    generated = 0
    freq_index = 0
    freq_list = list(freq_set)
    while generated < (freq_set_size*3):
        selected_key = freq_list[freq_index]
        file.write("key{0}\n".format(selected_key))
        freq_index += 1
        if freq_index >= freq_set_size:
            freq_index = 0
        generated += 1
    scan_size = 250
    scan_step = 0
    current_phase = "SCAN"
    total_index = 0
    while generated < key_count:
        if scan_step >= scan_size:
            scan_step = 0
            if current_phase == "SCAN":
                current_phase = "FREQ"
            else:
                current_phase = "SCAN"
        if current_phase == "SCAN":
            selected_key = available_keys[total_index]
            file.write("key{0}\n".format(selected_key))
            total_index += 1
            if total_index >= len(available_keys):
                total_index = 0
        else:
            selected_key = freq_list[freq_index]
            file.write("key{0}\n".format(selected_key))
            freq_index += 1
            if freq_index >= freq_set_size:
                freq_index = 0
        generated += 1
        scan_step += 1


def generate_lfu_keys(file, key_count, max_keyval):
    available_keys = range(0, max_keyval)
    freq_set_size = 200
    freq_set = set([])
    while len(freq_set) < freq_set_size:
        freq_set.add(random.choice(available_keys))
    generated = 0
    freq_set_probability = 0.99
    freq_complement = list(set(available_keys) - freq_set)
    freq_set = list(freq_set)
    scan_run_length = 200
    while generated < key_count:
        if random.random() < freq_set_probability:
            selected_key = random.choice(freq_set)
            file.write("key{0}\n".format(selected_key))
            generated += 1
        else:
            run_length = 0
            while run_length < scan_run_length and generated < key_count:
                selected_key = random.choice(freq_complement)
                file.write("key{0}\n".format(selected_key))
                generated += 1
                run_length += 1


def generate_lcr_keys(file, key_count, max_keyval):
    freq_set_size = 250
    cost_set_size = 250
    all_available_keys = range(0, max_keyval)
    freq_avail_keys = range(0, 1000)
    cost_avail_keys = range(int(max_keyval - (max_keyval/10)), max_keyval)
    freq_set = set([])
    while len(freq_set) < freq_set_size:
        freq_set.add(random.choice(freq_avail_keys))
    cost_set = set([])
    while len(cost_set) < cost_set_size:
        cost_set.add(random.choice(cost_avail_keys))
    complement_set = list((set(all_available_keys) - freq_set) - cost_set)
    freq_set = list(freq_set)
    cost_set = list(cost_set)
    scan_run_length = 200
    freq_set_prob = 0.495
    cost_set_prob = 0.495
    cost_set_threshold = freq_set_prob + cost_set_prob
    generated = 0
    while generated < key_count:
        r_val = random.random()
        if r_val < freq_set_prob:
            print("FREQ")
            selected_key = random.choice(freq_set)
            file.write("key{0}\n".format(selected_key))
            generated += 1
        elif r_val < cost_set_threshold:
            print("COST")
            selected_key = random.choice(cost_set)
            file.write("key{0}\n".format(selected_key))
            generated += 1
        else:
            run_length = 0
            while run_length < scan_run_length and generated < key_count:
                print("RUN")
                selected_key = random.choice(complement_set)
                file.write("key{0}\n".format(selected_key))
                generated += 1
                run_length += 1
    print(cost_set)
    print(freq_set)


def parse_arguments(args=None):
    parser = argparse.ArgumentParser(prog="calecar_key_gen")
    parser.add_argument("--cache_type", type=str)
    parser.add_argument("--output_filename", type=str)
    parser.add_argument("--output_count", type=int)
    parser.add_argument("--max_key_val", type=int)
    if args is None:
        parser.parse_args()  # will operate on sys.argv
    return parser.parse_args(args)


if __name__ == "__main__":
    args = parse_arguments()
    with open(args.output_filename, "w+") as f:
        if args.cache_type == "LRU":
            generate_lru_keys(f, args.output_count, args.max_key_val)
        elif args.cache_type == "LFU":
            generate_lfu_keys(f, args.output_count, args.max_key_val)
        elif args.cache_type == "LCR":
            generate_lcr_keys(f, args.output_count, args.max_key_val)
        elif args.cache_type == "LFU_SCAN":
            generate_lfu_scan_keys(f, args.output_count, args.max_key_val)
        elif args.cache_type == "LCR_SCAN":
            generate_lcr_scan_keys(f, args.output_count, args.max_key_val)
