# CVC word list generator project
CVC stand for Consonent/Vowel/Consonent

this project aim is to generate group of sets of CVC words.

we used phonetic alphabet
the langugae we used for this project was hebrew.

we cover limited set of CVC words as we want to use words which have actual meaning.

the subset for our project contain 20 consonant and 5 vowls

the collected around 450 valid words and figure out their usage frequency in the language

the requirement on the set are as follow
* a set must contain 10 CVC words
* a consonant cannot appear twice in the same set
* a vowel must appear twice

the requirement on the group are as follow
* each set must be balanced frequency wise
* words must appear only in one set within the group
* the group must have 20 sets

the solution was to go over all the permutations of the words and try to find a valid permutation

