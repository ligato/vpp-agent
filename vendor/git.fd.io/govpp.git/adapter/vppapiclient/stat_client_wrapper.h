// Copyright (c) 2018 Cisco and/or its affiliates.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at:
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

#ifndef included_stat_client_wrapper_h
#define included_stat_client_wrapper_h

#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include <arpa/inet.h>

#include <vpp-api/client/stat_client.h> // VPP has to be installed!

// The stat_client.h defines its version using two macros:
// 	STAT_VERSION_MAJOR - for major version
// 	STAT_VERSION_MINOR - for minor version
// both were introduced in VPP 19.04 (not on release, later on stable/1904)
// https://github.com/FDio/vpp/commit/1cb333cdf5ce26557233c5bdb5a18738cb6e1e2c

// Name vector directory type was introduced in VPP 19.04
#if STAT_VERSION_MAJOR >= 1 && STAT_VERSION_MINOR >= 1
	#define SUPPORTS_NAME_VECTOR // VPP 19.04 is required!
#endif

static int
govpp_stat_connect(char *socket_name)
{
	return stat_segment_connect(socket_name);
}

static void
govpp_stat_disconnect()
{
    stat_segment_disconnect();
}

static uint32_t*
govpp_stat_segment_ls(uint8_t **pattern)
{
	return stat_segment_ls(pattern);
}

static int
govpp_stat_segment_vec_len(void *vec)
{
	return stat_segment_vec_len(vec);
}

static void
govpp_stat_segment_vec_free(void *vec)
{
	stat_segment_vec_free(vec);
}

static char*
govpp_stat_segment_dir_index_to_name(uint32_t *dir, uint32_t index)
{
	return stat_segment_index_to_name(dir[index]);
}

static stat_segment_data_t*
govpp_stat_segment_dump(uint32_t *counter_vec)
{
	return stat_segment_dump(counter_vec);
}

static stat_segment_data_t
govpp_stat_segment_dump_index(stat_segment_data_t *data, int index)
{
	return data[index];
}

static int
govpp_stat_segment_data_type(stat_segment_data_t *data)
{
	return data->type;
}

static double
govpp_stat_segment_data_get_scalar_value(stat_segment_data_t *data)
{
	return data->scalar_value;
}

static double
govpp_stat_segment_data_get_error_value(stat_segment_data_t *data)
{
	return data->error_value;
}

static uint64_t**
govpp_stat_segment_data_get_simple_counter(stat_segment_data_t *data)
{
	return data->simple_counter_vec;
}

static uint64_t*
govpp_stat_segment_data_get_simple_counter_index(stat_segment_data_t *data, int index)
{
	return data->simple_counter_vec[index];
}

static uint64_t
govpp_stat_segment_data_get_simple_counter_index_value(stat_segment_data_t *data, int index, int index2)
{
	return data->simple_counter_vec[index][index2];
}

static vlib_counter_t**
govpp_stat_segment_data_get_combined_counter(stat_segment_data_t *data)
{
	return data->combined_counter_vec;
}

static vlib_counter_t*
govpp_stat_segment_data_get_combined_counter_index(stat_segment_data_t *data, int index)
{
	return data->combined_counter_vec[index];
}

static uint64_t
govpp_stat_segment_data_get_combined_counter_index_packets(stat_segment_data_t *data, int index, int index2)
{
	return data->combined_counter_vec[index][index2].packets;
}

static uint64_t
govpp_stat_segment_data_get_combined_counter_index_bytes(stat_segment_data_t *data, int index, int index2)
{
	return data->combined_counter_vec[index][index2].bytes;
}

static uint8_t**
govpp_stat_segment_data_get_name_vector(stat_segment_data_t *data)
{
#ifdef SUPPORTS_NAME_VECTOR
	return data->name_vector; // VPP 19.04 is required!
#else
	return 0;
#endif
}

static char*
govpp_stat_segment_data_get_name_vector_index(stat_segment_data_t *data, int index)
{
#ifdef SUPPORTS_NAME_VECTOR
	return data->name_vector[index]; // VPP 19.04 is required!
#else
	return 0;
#endif
}

static void
govpp_stat_segment_data_free(stat_segment_data_t *data)
{
	stat_segment_data_free(data);
}

static uint8_t**
govpp_stat_segment_string_vector(uint8_t ** string_vector, char *string)
{
	return stat_segment_string_vector(string_vector, string);
}

#endif /* included_stat_client_wrapper_h */
