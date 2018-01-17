cmake_minimum_required (VERSION 3.0)
project (jpeg-archive)

enable_testing ()

include (ExternalProject)

add_subdirectory (src)
add_subdirectory (test)

include_directories (SYSTEM ${MOZJPEG_INCLUDE_DIR})

add_executable (jpeg-compare jpeg-compare.c)
target_link_libraries (jpeg-compare jpegarchive IQA MOZJPEG -lm)

add_executable (jpeg-hash jpeg-hash.c)
target_link_libraries (jpeg-hash jpegarchive IQA MOZJPEG -lm)

add_executable (jpeg-recompress jpeg-recompress.c)
target_link_libraries (jpeg-recompress jpegarchive IQA MOZJPEG -lm)

find_program (GO_PROGRAM NAMES go)

add_custom_target (
    jpeg-archive-inplace-app ALL
    ${CMAKE_COMMAND}
        -E env
        CGO_CFLAGS="-I$<JOIN:$<TARGET_PROPERTY:jpeg-archive-inplace-app,INCLUDE_DIRECTORIES>, -I>"
        CGO_LDFLAGS='$<TARGET_FILE:jpegarchive> $<TARGET_FILE:IQA> $<TARGET_FILE:MOZJPEG> -lm'
        ${GO_PROGRAM}
            build
            -o ${CMAKE_CURRENT_BINARY_DIR}/jpeg-archive-inplace${CMAKE_EXECUTABLE_SUFFIX}
            -i
            ${CMAKE_CURRENT_SOURCE_DIR}/jpeg-archive-inplace.go
    BYPRODUCTS ${CMAKE_CURRENT_BINARY_DIR}/jpeg-archive-inplace${CMAKE_EXECUTABLE_SUFFIX}
    COMMENT go build jpeg-archive-inplace.go
    DEPENDS jpegarchive IQA MOZJPEG
    SOURCES jpeg-archive-inplace.go
)

set_property (
    DIRECTORY
    APPEND PROPERTY ADDITIONAL_MAKE_CLEAN_FILES
        ${CMAKE_CURRENT_BINARY_DIR}/jpeg-archive-inplace${CMAKE_EXECUTABLE_SUFFIX}
)

install (
    TARGETS jpeg-compare jpeg-hash jpeg-recompress jpeg-archive-inplace-app
    DESTINATION bin
)
