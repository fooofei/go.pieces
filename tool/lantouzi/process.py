#coding=utf-8

from typing import List
from datetime import timedelta
from datetime import datetime

def main():
    path = ""
    path2 = ""

    values = []
    with open(path, "rb") as fr:
        for line in fr:
            line = line.decode("utf-8")
            line = line.rstrip()
            fields = line.split(",")
            if len(fields) > 4:
                fs = ["-".join([fields[0], fields[1], fields[2]])]
                fs.append(fields[3])
                fs.append(fields[4])
                values.append(fs)

    values_map = {}
    for value in values:
        k = value[0]
        values_map[k] = value

    end_day = from_str("2019-12-08T00:00:00.00000")
    # sorted by multi keys
    # values.sort(key=lambda e:(e[0],e[1],e[2]), reverse=True)
    with open(path2, "wb") as fw:
        while len(values_map) > 0:
            k = time_key(end_day)
            if k in values_map:
                value = values_map[k]
                fw.write(",".join(value).encode("utf-8"))
                fw.write("\n".encode("utf-8"))
                values_map.pop(k, None)
            else:
                s = "%d-%02d-%02d,," % (end_day.year, end_day.month, end_day.day)
                fw.write(s.encode("utf-8"))
                fw.write("\n".encode("utf-8"))
            end_day = end_day - timedelta(days=1)


def time_key(t):
    return f"{t.date()}"

def from_str(dt_str):
    '''
    a = datetime.now()
    b = from_str(str(a.isoformat()))
    :param dt_str:
    :return:
    '''
    dt, _, us = dt_str.partition(".")
    dt = datetime.strptime(dt, "%Y-%m-%dT%H:%M:%S")
    us = int(us.rstrip("Z"), 10)
    return dt + timedelta(microseconds=us)

if __name__ == '__main__':
    main()

