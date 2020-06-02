echo "Test#1: "
ls > ../test.txt
./my-ls-1 -1 --nocolor> ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#2: "
ls utils.go > ../test.txt
./my-ls-1 -1 utils.go --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#3: "
ls dir > ../test.txt
./my-ls-1 -1 dir --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#4: "
ls -l > ../test.txt
./my-ls-1 -l --nocolor> ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#5: "
ls -l utils.go > ../test.txt
./my-ls-1 -l utils.go --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#6: "
ls -l dir > ../test.txt
./my-ls-1 -l dir --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#7: "
ls -l /usr/bin > ../test.txt
./my-ls-1 -l /usr/bin --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#8: "
ls -R ../ascii-art > ../test.txt
./my-ls-1 -R1 ../ascii-art --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#9: "
ls -a > ../test.txt
./my-ls-1 -a1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#10: "
ls -r > ../test.txt
./my-ls-1 -r1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#11: "
ls -t > ../test.txt
./my-ls-1 -t1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#12: "
ls -la > ../test.txt
./my-ls-1 -la --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#13: "
ls -l -t ../ascii-art > ../test.txt
./my-ls-1 -l -t ../ascii-art --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#14: "
ls -lRr ../ascii-art > ../test.txt
./my-ls-1 -lRr ../ascii-art --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#15: "
ls -l ../ascii-art -a utils.go > ../test.txt
./my-ls-1 -l ../ascii-art -a utils.go --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#16: "
ls -lR ../ascii-art///ascii-art/// ../ascii-art-web/src/ > ../test.txt
./my-ls-1 -lR ../ascii-art///ascii-art/// ../ascii-art-web/src/ --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

# echo "Test#17: "
# ls -la /dev > ../test.txt
# ./my-ls-1 -la /dev --nocolor > ../test2.txt
# diff ../test.txt ../test2.txt

echo "Test#18: "
ls -alRrt ../ascii-art > ../test.txt
./my-ls-1 -alRrt ../ascii-art --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#19: "
ls - > ../test.txt
./my-ls-1 - -1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#20: "
ls test2/ > ../test.txt
./my-ls-1 test2/ --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#21: "
ls test2 > ../test.txt
./my-ls-1 test2 -1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#22: "
ls test/ > ../test.txt
./my-ls-1 test/ -1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

echo "Test#23: "
ls "test" > ../test.txt
./my-ls-1 "test" -1 --nocolor > ../test2.txt
diff ../test.txt ../test2.txt

