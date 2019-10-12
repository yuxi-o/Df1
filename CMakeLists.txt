cmake_minimum_required(VERSION 2.6)

SET(CMAKE_BUILD_TYPE "Release")
SET(CMAKE_CXX_FLAGS_DEBUG "$ENV{CXXFLAGS} -O0 -Wall -g2 -ggdb")
SET(CMAKE_CXX_FLAGS_RELEASE "$ENV{CXXFLAGS} -O3 -Wall")
		
#Déclaration du projet
project(df1d)
set(EXECUTABLE_OUTPUT_PATH bin/${CMAKE_BUILD_TYPE})

#Génération de la liste des fichiers sources
file(
	GLOB_RECURSE
	source_files
	src/*
)

#Déclaration de l'exécutable
add_executable(
	df1d
	${source_files}
)
