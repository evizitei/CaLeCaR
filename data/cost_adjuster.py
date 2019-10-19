import argparse
import csv


def produce_new_costfile(rf, wf, direction, factor):
    reader = csv.reader(rf)
    writer = csv.writer(wf)
    COST_BASE = 1000
    for row in reader:
        old_cost = int(row[2])
        cost_delta = (old_cost - COST_BASE)
        new_cost = old_cost
        if direction == "UP":
            new_cost = COST_BASE + int(cost_delta * factor)
        elif direction == "DOWN":
            new_cost = COST_BASE + int(cost_delta / factor)
        else:
            raise RuntimeError("No such direction: " + direction)
        output_row = [row[0], row[1], new_cost]
        writer.writerow(output_row)


def parse_arguments(args=None):
    parser = argparse.ArgumentParser(prog="cost_adjuster")
    parser.add_argument("--input_filename", type=str)
    parser.add_argument("--direction", type=str)
    parser.add_argument("--output_filename", type=str)
    parser.add_argument("--factor", type=int)
    if args is None:
        parser.parse_args()  # will operate on sys.argv
    return parser.parse_args(args)


if __name__ == "__main__":
    args = parse_arguments()
    with open(args.input_filename, "r") as rf:
        with open(args.output_filename, "w+") as wf:
            produce_new_costfile(rf, wf, args.direction, args.factor)
