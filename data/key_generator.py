import argparse
import numpy as np
import random


def generate_lru_keys(file, key_count):
    mean = 5000
    generated = 0

    def sample_int_in_range(mu, minval=0, maxval=9999):
        standard_deviation = 8
        key_sample = np.random.normal(loc=mu, scale=standard_deviation)
        return int(min(maxval, max(minval, np.round(key_sample))))

    while generated < key_count:
        key = sample_int_in_range(mean)
        file.write("key{0}\n".format(key))
        # target_range_mean = mean + ((5000.0 - mean)/100000.0)
        mean = sample_int_in_range(mean, minval=75, maxval=9925)
        generated += 1


def generate_lfu_keys(file, key_count):
    available_keys = range(0, 10000)
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


def generate_lcr_keys(file, key_count):
    freq_set_size = 250
    cost_set_size = 250
    all_available_keys = range(0, 10000)
    freq_avail_keys = range(0, 2000)
    cost_avail_keys = range(8000, 10000)
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
    if args is None:
        parser.parse_args()  # will operate on sys.argv
    return parser.parse_args(args)


if __name__ == "__main__":
    args = parse_arguments()
    with open(args.output_filename, "w+") as f:
        if args.cache_type == "LRU":
            generate_lru_keys(f, args.output_count)
        elif args.cache_type == "LFU":
            generate_lfu_keys(f, args.output_count)
        elif args.cache_type == "LCR":
            generate_lcr_keys(f, args.output_count)
