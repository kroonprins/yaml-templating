Template:

```yaml
~[#merge .myobj]: {}
aa:
  bb[#append .myobj][#append .myobjlist]:
  - x: ii
  - p: o
  cc[#merge .myobj]:
    ee: dsf
  xx[#if .do]:
  - a
  - b
bb[#append .mylist]: []
re[#append .myobjlist]: []
iq[#prepend .myobj][#append .myobjlist][#insert .myobj2 2]:
- x: ii
- p: o
cc[#merge .myobj]: {}
xv:
  abc[#value .mystring]: \1
  def[#value .mystring]: abc \1 def
  ghi[#value .mystring .mysecondstring .myobj.ccc]: \1 and \2 and \3
dsf[#repeat .myobjlist]:
- a[#value $item.rrr]: \1
  b[#value $item.yyy]: \1
  c: c
xx:
  aa[#value .myobj.aaa]: \1
  bb[#value .myobj.ccc]: \1
  cc: cc
yy[#include include.yaml]: {}
~[#include include.yaml]: {}
tt:
  cc[#merge .myobj]: {}
  dd[#merge .myobj]: {}
  ee:
    cc[#merge .myobj][#merge .myobj2]: {}
    dd[#merge .myobj .myobj2]:
      yo: dela
      it: iiiee

```

Environment:
```yaml
mystring: yo
mysecondstring: mama
myobj:
  aaa: bbb
  ccc: ddd
myobj2:
  rrr: hhh
  fff: ttt
mylist:
  - zzz
  - yyy
  - xxx
myobjlist:
  - rrr: sss
    yyy: iii
  - rrr: uuu
    yyy: jjj
do: false
```

Result:

```yaml
aa:
  bb:
  - x: ii
  - p: o
  - aaa: bbb
    ccc: ddd
  - rrr: sss
    yyy: iii
  - rrr: uuu
    yyy: jjj
  cc:
    aaa: bbb
    ccc: ddd
    ee: dsf
  xx:
  - a
  - b
aaa: bbb
bb:
- aaa: bbb
  ccc: ddd
- zzz
- fff: ttt
  rrr: hhh
- yyy
- xxx
- rrr: sss
  yyy: iii
- rrr: uuu
  yyy: jjj
cc:
  aaa: bbb
  ccc: ddd
ccc: ddd
dsf:
- a: sss
  b: iii
  c: c
- a: uuu
  b: jjj
  c: c
include: hooray
op:
  aaa: bbb
  ccc: ddd
re:
- rrr: sss
  yyy: iii
- rrr: uuu
  yyy: jjj
tt:
  cc:
    aaa: bbb
    ccc: ddd
  dd:
    aaa: bbb
    ccc: ddd
  ee:
    cc:
      aaa: bbb
      ccc: ddd
      fff: ttt
      rrr: hhh
    dd:
      aaa: bbb
      ccc: ddd
      fff: ttt
      it: iiiee
      rrr: hhh
      yo: dela
xv:
  abc: yo
  def: abc yo def
  ghi: yo and mama and ddd
xx:
  aa: bbb
  bb: ddd
  cc: cc
yy:
  include: hooray
  op:
    aaa: bbb
    ccc: ddd
```
