import argparse
import numpy as np


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
        #target_range_mean = mean + ((5000.0 - mean)/100000.0)
        mean = sample_int_in_range(mean, minval=75, maxval=9925)
        generated += 1


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
