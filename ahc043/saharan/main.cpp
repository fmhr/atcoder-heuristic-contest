#if ONLINE_JUDGE
#define NDEBUG
//#pragma GCC target("avx2")
#pragma GCC optimize("O3")
#pragma GCC optimize("unroll-loops")
#else
#undef NDEBUG
#endif

#include <algorithm>
#include <array>
#include <bit>
#include <bitset>
#include <cassert>
#include <chrono>
#include <cmath>
#include <complex>
#include <concepts>
#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <deque>
#include <fstream>
#include <functional>
#include <initializer_list>
#include <iostream>
#include <limits>
#include <map>
#include <memory>
#include <mutex>
#include <numbers>
#include <numeric>
#include <optional>
#include <ostream>
#include <queue>
#include <ranges>
#include <set>
#include <sstream>
#include <stack>
#include <string>
#include <thread>
#include <tuple>
#include <unordered_map>
#include <unordered_set>
#include <utility>
#include <vector>

// note: a huge library begins, the main part starts around 2700

namespace shr {
	namespace basic {
		using namespace std;
		using uchar = unsigned char;
		using uint = unsigned int;
		using ushort = unsigned short;
		using ull = unsigned long long;
		using ll = long long;
		using pii = pair<int, int>;
		using pdi = pair<double, int>;
#define fail(reason) assert(std::make_pair(reason, false).second)

		template <class T>
		concept printable = requires(T t, ostream& out) { out << t; };

		int len(const string& a) {
			return (int) a.length();
		}

		template <class T>
		int len(const vector<T>& a) {
			return (int) a.size();
		}

		template <class T>
		int len(const set<T>& a) {
			return (int) a.size();
		}

		template <class T>
		int len(const unordered_set<T>& a) {
			return (int) a.size();
		}

		template <class T>
		int len(const deque<T>& a) {
			return (int) a.size();
		}

		template <class T>
		int len(const queue<T>& a) {
			return (int) a.size();
		}

		template <class T>
		int len(const priority_queue<T>& a) {
			return (int) a.size();
		}

		template <class T, int N>
		int len(const T (&a)[N]) {
			return N;
		}

		template <class T, int N>
		void clear_with(T (&a)[N], int val) {
			memset(a, val, sizeof(a));
		}

		template <totally_ordered T>
		bool update_min(T& current_min, T candidate) {
			if (candidate < current_min) {
				current_min = candidate;
				return true;
			}
			return false;
		}

		template <totally_ordered T>
		bool update_min_eq(T& current_min, T candidate) {
			if (candidate <= current_min) {
				current_min = candidate;
				return true;
			}
			return false;
		}

		template <totally_ordered T>
		bool update_max(T& current_max, T candidate) {
			if (candidate > current_max) {
				current_max = candidate;
				return true;
			}
			return false;
		}

		template <totally_ordered T>
		bool update_max_eq(T& current_max, T candidate) {
			if (candidate >= current_max) {
				current_max = candidate;
				return true;
			}
			return false;
		}

		template <class T>
		string tos(T a) {
			return to_string(a);
		}

		template <printable T>
		string tos(T a) {
			ostringstream os;
			os << a;
			return os.str();
		}

		template <class T>
		requires(!printable<T>)
		string tos(T a) {
			ostringstream os;
			bool first = true;
			auto from = a.begin();
			auto until = a.end();
			os << "{";
			while (from != until) {
				if (first) {
					first = false;
				} else {
					os << ", ";
				}
				os << tos(*from);
				from++;
			}
			os << "}";
			return os.str();
		}

		constexpr double linearstep(double edge0, double edge1, double t) {
			return clamp((t - edge0) / (edge1 - edge0), 0.0, 1.0);
		}

		constexpr double smoothstep(double edge0, double edge1, double t) {
			t = linearstep(edge0, edge1, t);
			return t * t * (3 - 2 * t);
		}

		double exp_interp(double from, double to, double t) {
			return pow(from, 1 - t) * pow(to, t);
		}

		template <ranges::range Range, class F>
		auto mapped(Range& as, F f) {
			using T = ranges::range_value_t<Range>;
			using U = invoke_result_t<F, T&>;
			vector<U> res;
			for (auto& a : as) {
				res.push_back(f(a));
			}
			return res;
		}

		template <ranges::range Range, class F>
		auto mapped(const Range& as, F f) {
			using T = ranges::range_value_t<Range>;
			using U = invoke_result_t<F, const T&>;
			vector<U> res;
			for (auto& a : as) {
				res.push_back(f(a));
			}
			return res;
		}

		bool msg(bool value, const string& message) {
			return value;
		}
	} // namespace basic
	using namespace basic;

	namespace timer {
		double time_scale = 1.0;

		// return in ms
		int timer(bool reset = false) {
			static auto st = chrono::system_clock::now();
			if (reset) {
				st = chrono::system_clock::now();
				return 0;
			} else {
				auto en = chrono::system_clock::now();
				int elapsed = (int) chrono::duration_cast<chrono::milliseconds>(en - st).count();
				return (int) round(elapsed / time_scale);
			}
		}
	} // namespace timer

	namespace tracer {
		bool debug = true;

		template <class T>
		concept is_pair = requires(T t) {
			t.first;
			t.second;
		};

		template <class T>
		concept has_str = requires(T t) {
			{ t.str() } -> convertible_to<string>;
		};

		template <printable T>
		void tracen(T&& t) {
			if (!debug)
				return;
			cerr << t;
		}

		template <class T>
		requires(has_str<T> && !printable<T>)
		void tracen(T&& t) {
			if (!debug)
				return;
			cerr << t.str();
		}

		template <class T, class U>
		void tracen(pair<T, U>& t) { // <- ?????????????????????? need this for
			// trace(<iterable of pairs>)
			if (!debug)
				return;
			cerr << "(";
			tracen(t.first);
			cerr << ", ";
			tracen(t.second);
			cerr << ")";
		}

		template <class T, class U>
		void tracen(pair<T, U>&& t) { // <- ?????????????????????? need this for
			// trace(make_pair(1, 2))
			if (!debug)
				return;
			cerr << "(";
			tracen(t.first);
			cerr << ", ";
			tracen(t.second);
			cerr << ")";
		}

		template <class T>
		requires(!printable<T>)
		void tracen(T&& t) {
			if (!debug)
				return;
			bool first = true;
			auto from = t.begin();
			auto until = t.end();
			cerr << "{";
			while (from != until) {
				if (first) {
					first = false;
				} else {
					cerr << ", ";
				}
				tracen(*from);
				from++;
			}
			cerr << "}";
		}

		template <class T, int N>
		requires(!same_as<decay_t<T>, char>)
		void tracen(T (&a)[N]) {
			if (!debug)
				return;
			cerr << "{";
			for (int i = 0; i < N; i++) {
				if (i > 0)
					cerr << ", ";
				tracen(a[i]);
			}
			cerr << "}";
		}

		template <class T1, class T2, class... Rest>
		void tracen(T1&& t1, T2&& t2, Rest&&... rest) {
			if (!debug)
				return;
			tracen(std::forward<T1>(t1));
			tracen(std::forward<T2>(t2), std::forward<Rest>(rest)...);
		}

		void trace() {
			if (!debug)
				return;
			cerr << endl;
		}

		template <class T, class... Rest>
		void trace(T&& t, Rest&&... rest) {
			if (!debug)
				return;
			tracen(std::forward<T>(t), std::forward<Rest>(rest)...);
			cerr << endl;
		}

		template <class T>
		requires(!printable<T>)
		void trace2d(T&& t, int h, int w) {
			if (!debug)
				return;
			bool first = true;
			auto from = t.begin();
			auto until = t.end();
			for (int i = 0; i < h; i++) {
				for (int j = 0; j < w; j++) {
					if (j > 0)
						tracen(" ");
					tracen(*from);
					from++;
					if (j == w - 1)
						trace();
				}
			}
		}

		template <class T, int N>
		requires(!same_as<decay_t<T>, char>)
		void trace2d(T (&a)[N], int h, int w) {
			if (!debug)
				return;
			int idx = 0;
			for (int i = 0; i < h; i++) {
				for (int j = 0; j < w; j++) {
					if (j > 0)
						tracen(" ");
					tracen(a[idx]);
					idx++;
					if (j == w - 1)
						trace();
				}
			}
		}
	} // namespace tracer
	using namespace tracer;

	namespace random {
		class rngen {
		public:
			rngen() {
			}

			// to avoid bugs
			rngen(const rngen&) = delete;

			rngen& operator=(const rngen&&) = delete;

			rngen(int s) {
				seed(s);
			}

			ull get_state() {
				return state;
			}

			void set_state(ull state) {
				this->state = state;
			}

			void seed(int s) {
				state = s + INCR;
				next32();
			}

			int next_int() {
				return next31();
			}

			int next_int(int mod) {
				assert(mod > 0);
				return (int) ((ull) next31() * mod >> 31);
			}

			int next_int(int min, int max) {
				return min + next_int(max - min + 1);
			}

			uint next_uint() {
				return next32();
			}

			ull next_ull() {
				return (ull) next32() << 32 | next32();
			}

			double next_float() {
				return (double) next31() / 0x80000000;
			}

			double next_float(double min, double max) {
				return min + next_float() * (max - min);
			}

			double next_normal() {
				return sqrt(-2 * log(next_float())) * cos(6.283185307179586 * next_float());
			}

			double next_normal(double mean, double sigma) {
				return mean + next_normal() * sigma;
			}

		private:
			static constexpr ull MULT = 0x8b46ad15ae59daadull;
			static constexpr ull INCR = 0xf51827be20401689ull;
			ull state = (ull) chrono::duration_cast<chrono::nanoseconds>(
				chrono::system_clock::now().time_since_epoch())
							.count();

			uint next32() {
				uint r = (uint) (state >> 59);
				state = state * MULT + INCR;
				state ^= state >> 18;
				uint t = (uint) (state >> 27);
				return t >> r | t << (-r & 31);
			}

			int next31() {
				return (int) (next32() & 0x7fffffff);
			}
		};

		void random_permutation(int* a, int n, rngen& rng) {
			assert(n >= 0);
			if (n == 0)
				return;
			a[0] = 0;
			for (int i = 1; i < n; i++) {
				a[i] = i;
				swap(a[i], a[rng.next_int(i + 1)]);
			}
		}

		template <class RandomAccessContainer>
		void shuffle(RandomAccessContainer& c, rngen& rng) {
			int n = len(c);
			for (int i = 1; i < n; i++) {
				swap(c[i], c[rng.next_int(i + 1)]);
			}
		}

		template <class RandomAccessContainer>
		auto& random_pick(RandomAccessContainer& c, rngen& rng) {
			return c[rng.next_int(len(c))];
		}

		template <class RandomAccessContainer>
		const auto& random_pick(const RandomAccessContainer& c, rngen& rng) {
			return c[rng.next_int(len(c))];
		}

		template <class T, int N>
		T& random_pick(T (&c)[N], rngen& rng) {
			return c[rng.next_int(N)];
		}

		template <class T, int N>
		const T& random_pick(const T (&c)[N], rngen& rng) {
			return c[rng.next_int(N)];
		}
	} // namespace random
	using namespace random;

	namespace ds {
		// random access: O(1)
		// push: O(1)
		// insert: n/a
		// erase by position: O(1)
		// erase by element: n/a
		// max size: fixed
		template <class T, int N>
		class fast_vector {
		private:
			T data[N];
			int num = 0;

		public:
			using iterator = T*;
			using const_iterator = const T*;

			iterator begin() {
				return data;
			}

			iterator end() {
				return data + num;
			}

			const_iterator begin() const {
				return data;
			}

			const_iterator end() const {
				return data + num;
			}

			void push_back(T a) {
				assert(num < N);
				data[num++] = a;
			}

			template <class... Args>
			T& emplace_back(Args&&... args) {
				push_back(T(std::forward<Args>(args)...));
				return data[num - 1];
			}

			void erase(iterator where) {
				assert(where >= begin() && where < end());
				*where = data[--num];
			}

			void pop_back() {
				assert(num > 0);
				num--;
			}

			void clear() {
				num = 0;
			}

			T& operator[](int i) {
				assert(i >= 0 && i < num);
				return data[i];
			}

			const T& operator[](int i) const {
				assert(i >= 0 && i < num);
				return data[i];
			}

			int size() const {
				return num;
			}

			bool empty() const {
				return num == 0;
			}

			void copyFrom(const fast_vector<T, N>& other) {
				num = other.num;
				memcpy(data, other.data, sizeof(T) * num);
			}
		};

		// random access: O(1)
		// push: O(1)
		// insert: n/a
		// erase: n/a
		// reallocation: never
		template <class T, int UnitBits = 20>
		class increasing_vector {
		private:
			static constexpr int UNIT_SIZE = 1 << UnitBits;
			static constexpr int UNIT_MASK = UNIT_SIZE - 1;
			T** packs;
			int num_packs = 0;
			int num_total = 0;

		public:
			increasing_vector(const increasing_vector& vec) = delete;
			increasing_vector& operator=(const increasing_vector& vec) = delete;

			increasing_vector() : packs(new T*[65536]) {
			}

			~increasing_vector() {
				for (int i = 0; i < num_packs; i++) {
					delete[] packs[i];
				}
				delete[] packs;
			}

			T* next_pointer() {
				if ((num_total++ & UNIT_MASK) == 0) {
					packs[num_packs++] = new T[UNIT_SIZE];
				}
				return &(*this)[num_total - 1];
			}

			void push_back(T a) {
				*next_pointer() = a;
			}

			template <class... Args>
			void emplace_back(Args&&... args) {
				push_back(T(std::forward<Args>(args)...));
			}

			T& operator[](int i) {
				assert(i >= 0 && i < num_total);
				return packs[i >> UnitBits][i & UNIT_MASK];
			}

			const T& operator[](int i) const {
				assert(i >= 0 && i < num_total);
				return packs[i >> UnitBits][i & UNIT_MASK];
			}

			int size() const {
				return num_total;
			}
		};

		// random access: O(1)
		// insert: O(1)
		// erase: O(1)
		// check: O(1)
		// max value: fixed
		template <int N>
		class fast_iset {
		private:
			int data[N];
			int indices[N];
			int num = 0;

		public:
			using iterator = int*;
			using const_iterator = const int*;

			fast_iset() {
				memset(indices, -1, sizeof(indices));
			}

			iterator begin() {
				return data;
			}

			iterator end() {
				return data + num;
			}

			const_iterator begin() const {
				return data;
			}

			const_iterator end() const {
				return data + num;
			}

			bool insert(int a) {
				assert(a >= 0 && a < N);
				if (indices[a] != -1)
					return false;
				data[num] = a;
				indices[a] = num;
				num++;
				return true;
			}

			bool erase(int a) {
				assert(a >= 0 && a < N);
				int index = indices[a];
				if (index == -1)
					return false;
				assert(num > 0);
				indices[data[index] = data[--num]] = index;
				indices[a] = -1;
				return true;
			}

			void clear() {
				memset(indices, -1, sizeof(indices));
				num = 0;
			}

			bool contains(int a) const {
				return indices[a] != -1;
			}

			const int& operator[](int i) const {
				assert(i >= 0 && i < num);
				return data[i];
			}

			int size() const {
				return num;
			}

			bool empty() const {
				return num == 0;
			}
		};

		// insert: O(1)
		// get/set: O(1)
		// clear: O(1)
		// erase: n/a
		template <class T, int BucketBits = 20>
		class hash_imap {
		private:
			static constexpr int BUCKET_SIZE = 1 << BucketBits;
			static constexpr int BUCKET_MASK = BUCKET_SIZE - 1;

			ull* keys;
			T* values;
			ushort* access_time;
			ushort time = (ushort) -1;
			int num_elements = 0;
			int last_index = -1;
			ull last_key = -1;
			bool last_found = false;

		public:
			hash_imap()
				: keys(new ull[BUCKET_SIZE]), values(new T[BUCKET_SIZE]),
				  access_time(new ushort[BUCKET_SIZE]) {
			}

			~hash_imap() {
				delete[] keys;
				delete[] values;
				delete[] access_time;
			}

			hash_imap(const hash_imap& map)
				: keys(new ull[BUCKET_SIZE]), values(new T[BUCKET_SIZE]),
				  access_time(new ushort[BUCKET_SIZE]) {
				memcpy(keys, map.keys, sizeof(ull[BUCKET_SIZE]));
				memcpy(values, map.values,
					sizeof(T[BUCKET_SIZE])); // can be potentially dangerous?
				memcpy(access_time, map.access_time, sizeof(ushort[BUCKET_SIZE]));
				time = map.time;
				num_elements = map.num_elements;
				last_index = map.last_index;
				last_key = map.last_key;
				last_found = map.last_found;
			}

			hash_imap& operator=(const hash_imap& map) {
				if (this == &map)
					return *this;
				delete[] keys;
				delete[] values;
				delete[] access_time;
				keys = new ull[BUCKET_SIZE];
				values = new T[BUCKET_SIZE];
				access_time = new ushort[BUCKET_SIZE];
				memcpy(keys, map.keys, sizeof(ull[BUCKET_SIZE]));
				memcpy(values, map.values,
					sizeof(T[BUCKET_SIZE])); // can be potentially dangerous?
				memcpy(access_time, map.access_time, sizeof(ushort[BUCKET_SIZE]));
				time = map.time;
				num_elements = map.num_elements;
				last_index = map.last_index;
				last_key = map.last_key;
				last_found = map.last_found;
				return *this;
			}

			void clear() {
				num_elements = 0;
				last_found = false;
				last_index = -1;
				if (++time == 0) {
					memset(access_time, 0, sizeof(ushort[BUCKET_SIZE]));
					time = 1;
				}
			}

			bool access(ull key) {
				last_key = key;
				last_index = (int) (key & BUCKET_MASK);
				bool debug = false;
				while (true) {
					if (access_time[last_index] != time) {
						return last_found = false;
					} else if (keys[last_index] == key) {
						return last_found = true;
					}
					last_index = (last_index + 1) & BUCKET_MASK;
				}
			}

			T get() const {
				assert(last_found);
				return values[last_index];
			}

			void set(T value) {
				assert(last_index != -1);
				access_time[last_index] = time;
				keys[last_index] = last_key;
				values[last_index] = value;
				num_elements += !last_found;
				assert(num_elements < 0.85 * BUCKET_SIZE); // bucket size is too small
			}
		};

		// a bitset, but cooler than std::bitset
		template <int Size>
		class rich_bitset {
		private:
			using word = ull;
			static_assert(has_single_bit(sizeof(word)));
			static constexpr int WORD_SHIFT = std::countr_zero(8 * sizeof(word));
			static constexpr int WORD_SIZE = 1 << WORD_SHIFT;
			static constexpr int WORD_MASK = WORD_SIZE - 1;
			static constexpr int NUM_WORDS = (Size + WORD_SIZE - 1) / WORD_SIZE;
			static constexpr int LAST_WORD = NUM_WORDS - 1;
			static constexpr word LAST_WORD_MASK =
				(Size & WORD_MASK) == 0 ? word(-1) : (word(1) << (Size & WORD_MASK)) - 1;
#define REP_WORDS(i) for (int i = 0; i < NUM_WORDS; i++)
#define REP_INNER_WORDS(i) for (int i = 0; i < NUM_WORDS - 1; i++)
#define REP_WORDS_REV(i) for (int i = NUM_WORDS - 1; i >= 0; i--)
#define REP_INNER_WORDS_REV(i) for (int i = NUM_WORDS - 2; i >= 0; i--)

			// [LAST_WORD] [LAST_WORD - 1] [...] [1] [0]
			// <- higher bits              lower bits ->
			word data[NUM_WORDS];

			struct ref {
				rich_bitset<Size>& bs;
				const int pos;

				ref(rich_bitset<Size>& bs, int pos) : bs(bs), pos(pos) {
				}

				ref& operator=(bool val) {
					bs.set(pos, val);
					return *this;
				}

				operator bool() const {
					return bs.test(pos);
				}
			};

			void trim() {
				if constexpr ((Size & WORD_MASK) != 0) {
					data[LAST_WORD] &= LAST_WORD_MASK;
				}
			}

		public:
			rich_bitset(ull value = 0) {
				constexpr int BITS = sizeof(ull) * 8;
				for (int i = 0; i < (BITS + WORD_SIZE - 1) / WORD_SIZE; i++) {
					data[i] = value >> i * WORD_SIZE;
				}
				constexpr int OFFSET = (BITS + WORD_SIZE - 1) / WORD_SIZE;
				if constexpr (OFFSET < NUM_WORDS) {
					memset(data + OFFSET, 0, sizeof(word) * (NUM_WORDS - OFFSET));
				}
			}

			bool all() const {
				bool res = true;
				REP_INNER_WORDS(i) {
					res &= data[i] == word(-1);
				}
				res &= data[LAST_WORD] == LAST_WORD_MASK;
				return res;
			}

			bool none() const {
				bool res = true;
				REP_WORDS(i) {
					res &= data[i] == 0;
				}
				return res;
			}

			bool any() const {
				bool res = false;
				REP_WORDS(i) {
					res |= data[i] != 0;
				}
				return res;
			}

			int count() const {
				int res = 0;
				REP_WORDS(i) {
					res += popcount(data[i]);
				}
				return res;
			}

			int countr_zero() const {
				if constexpr (LAST_WORD == 0) {
					return std::countr_zero(word(data[LAST_WORD] | ~LAST_WORD_MASK));
				} else {
					int res = std::countr_zero(data[0]);
					int mask = -(res == WORD_SIZE); // continue adding if -1
					for (int i = 1; i < NUM_WORDS - 1; i++) {
						int count = std::countr_zero(data[i]);
						res += count & mask;
						mask &= -(count == WORD_SIZE);
					}
					int count = std::countr_zero(word(data[LAST_WORD] | ~LAST_WORD_MASK));
					res += count & mask;
					return res;
				}
			}

			int countl_zero() const {
				constexpr int LAST_WORD_SIZE = popcount(LAST_WORD_MASK);
				int res = std::countl_zero(word(~(~data[LAST_WORD] << (WORD_SIZE - LAST_WORD_SIZE))));
				int mask = -(res == LAST_WORD_SIZE); // continue adding if -1
				for (int i = NUM_WORDS - 2; i >= 0; i--) {
					int count = std::countl_zero(data[i]);
					res += count & mask;
					mask &= -(count == WORD_SIZE);
				}
				return res;
			}

			int countr_one() const {
				if constexpr (LAST_WORD == 0) {
					return std::countr_one(data[LAST_WORD]);
				} else {
					int res = std::countr_one(data[0]);
					int mask = -(res == WORD_SIZE); // continue adding if -1
					for (int i = 1; i < NUM_WORDS - 1; i++) {
						int count = std::countr_one(data[i]);
						res += count & mask;
						mask &= -(count == WORD_SIZE);
					}
					int count = std::countr_one(data[LAST_WORD]);
					res += count & mask;
					return res;
				}
			}

			int countl_one() const {
				constexpr int LAST_WORD_SIZE = popcount(LAST_WORD_MASK);
				int res = std::countl_one(word(data[LAST_WORD] << (WORD_SIZE - LAST_WORD_SIZE)));
				int mask = -(res == LAST_WORD_SIZE); // continue adding if -1
				for (int i = NUM_WORDS - 2; i >= 0; i--) {
					int count = std::countl_one(data[i]);
					res += count & mask;
					mask &= -(count == WORD_SIZE);
				}
				return res;
			}

			int size() const {
				return Size;
			}

			bool test(int pos) const {
				assert(pos >= 0 && pos < Size);
				return (data[pos >> WORD_SHIFT] >> (pos & WORD_MASK)) & 1;
			}

			uint to_uint() const {
				constexpr int BITS = sizeof(uint) * 8;
				for (int i = (BITS + WORD_SIZE - 1) / WORD_SIZE; i < NUM_WORDS; i++) {
					assert(("uint overflow", data[i] == 0));
				}
				if constexpr (WORD_SIZE > BITS) {
					assert(("uint overflow", (data[0] >> BITS) == 0));
				}
				uint res = (uint) data[0];
				for (int i = 1; i < (BITS + WORD_SIZE - 1) / WORD_SIZE && i < NUM_WORDS; i++) {
					res |= (uint) data[i] << i * WORD_SIZE;
				}
				return res;
			}

			ull to_ull() const {
				constexpr int BITS = sizeof(ull) * 8;
				for (int i = (BITS + WORD_SIZE - 1) / WORD_SIZE; i < NUM_WORDS; i++) {
					assert(("ull overflow", data[i] == 0));
				}
				if constexpr (WORD_SIZE > BITS) {
					assert(("ull overflow", (data[0] >> BITS) == 0));
				}
				ull res = (ull) data[0];
				for (int i = 1; i < (BITS + WORD_SIZE - 1) / WORD_SIZE && i < NUM_WORDS; i++) {
					res |= (ull) data[i] << i * WORD_SIZE;
				}
				return res;
			}

			rich_bitset& set(int pos, bool val = true) {
				assert(pos >= 0 && pos < Size);
				word bit = word(1) << (pos & WORD_MASK);
				if (val) {
					data[pos >> WORD_SHIFT] |= bit;
				} else {
					data[pos >> WORD_SHIFT] &= ~bit;
				}
				return *this;
			}

			rich_bitset& reset(int pos) {
				assert(pos >= 0 && pos < Size);
				return set(pos, false);
			}

			rich_bitset& flip(int pos) {
				assert(pos >= 0 && pos < Size);
				word bit = word(1) << (pos & WORD_MASK);
				data[pos >> WORD_SHIFT] ^= bit;
				return *this;
			}

			rich_bitset& set() {
				clear_with(data, -1);
				trim();
				return *this;
			}

			rich_bitset& reset() {
				clear_with(data, 0);
				return *this;
			}

			rich_bitset& flip() {
				REP_INNER_WORDS(i) {
					data[i] ^= word(-1);
				}
				data[LAST_WORD] ^= LAST_WORD_MASK;
				return *this;
			}

			word* words() {
				return data;
			}

			rich_bitset& operator&=(const rich_bitset& a) {
				REP_WORDS(i) {
					data[i] &= a.data[i];
				}
				return *this;
			}

			rich_bitset& operator|=(const rich_bitset& a) {
				REP_WORDS(i) {
					data[i] |= a.data[i];
				}
				return *this;
			}

			rich_bitset& operator^=(const rich_bitset& a) {
				REP_WORDS(i) {
					data[i] ^= a.data[i];
				}
				return *this;
			}

			rich_bitset& operator<<=(int amount) {
				assert(amount >= 0 && amount < Size);
				int nw = amount >> WORD_SHIFT;
				if (nw > 0) {
					REP_WORDS_REV(i) {
						data[i] = i - nw < 0 ? 0 : data[i - nw];
					}
				}
				int nb = amount & WORD_MASK;
				if (nb) {
					for (int i = NUM_WORDS - 1; i > 0; i--) {
						data[i] = data[i] << nb | data[i - 1] >> (WORD_SIZE - nb);
					}
					data[0] <<= nb;
				}
				trim();
				return *this;
			}

			rich_bitset& operator>>=(int amount) {
				assert(amount >= 0 && amount < Size);
				int nw = amount >> WORD_SHIFT;
				if (nw > 0) {
					REP_WORDS(i) {
						data[i] = i + nw >= NUM_WORDS ? 0 : data[i + nw];
					}
				}
				int nb = amount & WORD_MASK;
				if (nb) {
					REP_INNER_WORDS(i) {
						data[i] = data[i] >> nb | data[i + 1] << (WORD_SIZE - nb);
					}
					data[LAST_WORD] >>= nb;
				}
				return *this;
			}

			rich_bitset& operator+=(const rich_bitset& a) {
				word carry = 0;
				REP_WORDS(i) {
					word l = data[i];
					word r = a.data[i];
					word sum = l + r;
					data[i] = sum + carry;
					carry = (sum < l) | (data[i] < sum);
				}
				trim();
				return *this;
			}

			rich_bitset& operator-=(const rich_bitset& a) {
				word carry = 1;
				REP_WORDS(i) {
					word l = data[i];
					word r = ~a.data[i];
					word sum = l + r;
					data[i] = sum + carry;
					carry = (sum < l) | (data[i] < sum);
				}
				trim();
				return *this;
			}

			rich_bitset& operator++() {
				word carry = 1;
				REP_WORDS(i) {
					word l = data[i];
					data[i] = l + carry;
					carry = (data[i] < l);
				}
				trim();
				return *this;
			}

			rich_bitset operator++(int) {
				rich_bitset res = *this;
				operator++();
				return res;
			}

			rich_bitset& operator--() {
				word carry = 0;
				REP_WORDS(i) {
					word l = data[i];
					data[i] = l - 1 + carry;
					carry = (l | carry) != 0;
				}
				trim();
				return *this;
			}

			rich_bitset operator--(int) {
				rich_bitset res = *this;
				operator--();
				return res;
			}

			rich_bitset operator~() const {
				rich_bitset res = *this;
				res.flip();
				return res;
			}

			friend rich_bitset operator&(const rich_bitset& a, const rich_bitset& b) {
				rich_bitset res = a;
				res &= b;
				return res;
			}

			friend rich_bitset operator|(const rich_bitset& a, const rich_bitset& b) {
				rich_bitset res = a;
				res |= b;
				return res;
			}

			friend rich_bitset operator^(const rich_bitset& a, const rich_bitset& b) {
				rich_bitset res = a;
				res ^= b;
				return res;
			}

			friend rich_bitset operator<<(const rich_bitset& a, int amount) {
				rich_bitset res = a;
				res <<= amount;
				return res;
			}

			friend rich_bitset operator>>(const rich_bitset& a, int amount) {
				rich_bitset res = a;
				res >>= amount;
				return res;
			}

			friend rich_bitset operator+(const rich_bitset& a, const rich_bitset& b) {
				rich_bitset res = a;
				res += b;
				return res;
			}

			friend rich_bitset operator-(const rich_bitset& a, const rich_bitset& b) {
				rich_bitset res = a;
				res -= b;
				return res;
			}

			friend bool operator==(const rich_bitset& a, const rich_bitset& b) {
				return memcmp(a.data, b.data, sizeof(a.data)) == 0;
			}

			friend bool operator!=(const rich_bitset& a, const rich_bitset& b) {
				return memcmp(a.data, b.data, sizeof(a.data)) != 0;
			}

			friend int operator<=>(const rich_bitset& a, const rich_bitset& b) {
				REP_WORDS_REV(i) {
					if (a.data[i] != b.data[i])
						return a.data[i] < b.data[i] ? -1 : 1;
				}
				return 0;
			}

			ref operator[](int pos) {
				return {*this, pos};
			}

			bool operator[](int pos) const {
				return test(pos);
			}

			string str() const {
				ostringstream oss;
				oss << *this;
				return oss.str();
			}

			friend ostream& operator<<(ostream& out, const rich_bitset& bs) {
				for (int i = Size - 1; i >= 0; i--) {
					out << (bs.test(i) ? '1' : '0');
				}
				return out;
			}
#undef REP_WORDS
#undef REP_INNER_WORDS
#undef REP_WORDS_REV
#undef REP_INNER_WORDS_REV
		};

		template <class T>
		class easy_stack {
		public:
			vector<T> data;

			auto begin() {
				return data.begin();
			}

			auto end() {
				return data.end();
			}

			auto begin() const {
				return data.begin();
			}

			auto end() const {
				return data.end();
			}

			void clear() {
				data.clear();
			}

			void push_back(T a) {
				data.push_back(a);
			}

			template <class... Args>
			auto& emplace_back(Args&&... args) {
				return data.emplace_back(std::forward<Args>(args)...);
			}

			T pop_back() {
				T res = data.back();
				data.pop_back();
				return res;
			}

			T& back() {
				return data.back();
			}

			const T& back() const {
				return data.back();
			}

			int size() const {
				return (int) data.size();
			}

			bool empty() const {
				return data.empty();
			}
		};

		template <int N>
		int len(const fast_iset<N>& a) {
			return a.size();
		}

		template <class T, int N>
		int len(const fast_vector<T, N>& a) {
			return a.size();
		}

		template <class T, int BucketBits>
		int len(const hash_imap<T, BucketBits>& a) {
			return a.size();
		}

		template <class T>
		int len(const easy_stack<T>& a) {
			return a.size();
		}

		template <class T>
		requires(same_as<T, char> || same_as<T, short> || same_as<T, int>)
		struct int_vec2 {
			T i;
			T j;

			template <class U>
			constexpr int_vec2(int_vec2<U> a) : i(a.i), j(a.j) {
				assert(i == a.i);
				assert(j == a.j);
			}

			constexpr int_vec2() : i(0), j(0) {
			}

			constexpr int_vec2(T i, T j) : i(i), j(j) {
			}

			constexpr static int_vec2 dir(int index) {
				constexpr T DIRS[4][2] = {
					{-1, 0},
					{1, 0},
					{0, -1},
					{0, 1},
				};
				return {DIRS[index][0], DIRS[index][1]};
			}

			constexpr int dir_index() const {
				assert((i != 0) + (j != 0) == 1);
				return i < 0 ? 0 : i > 0 ? 1 : j < 0 ? 2 : 3;
			}

			constexpr int_vec2 rot(int sij, int num = 1) const {
				num &= 3;
				int_vec2 res = {i, j};
				while (num) {
					res = {sij - 1 - res.j, res.i};
					num--;
				}
				return res;
			}

			constexpr int_vec2 min(int_vec2 a) const {
				return {std::min(i, a.i), std::min(j, a.j)};
			}

			constexpr int_vec2 max(int_vec2 a) const {
				return {std::max(i, a.i), std::max(j, a.j)};
			}

			constexpr int_vec2 clamp(int_vec2 min, int_vec2 max) const {
				return {std::clamp(i, min.i, max.i), std::clamp(j, min.j, max.j)};
			}

			int_vec2 abs() const {
				return {std::abs(i), std::abs(j)};
			}

			// manhattan norm
			int mnorm() const {
				return std::abs(i) + std::abs(j);
			}

			// euclidean norm
			double enorm() const {
				return sqrt(lldot(*this));
			}

			constexpr int dot(int_vec2 a) const {
				return i * a.i + j * a.j;
			}

			constexpr ll lldot(int_vec2 a) const {
				return (ll) i * a.i + (ll) j * a.j;
			}

			constexpr int pack(int sij) const {
				return pack(sij, sij);
			}

			constexpr int pack(int si, int sj) const {
				assert(in_bounds(si, sj));
				return i * sj + j;
			}

			constexpr int pack_if_in_bounds(int sij) const {
				return pack_if_in_bounds(sij, sij);
			}

			constexpr int pack_if_in_bounds(int si, int sj) const {
				if (!in_bounds(si, sj))
					return -1;
				return i * sj + j;
			}

			constexpr static int_vec2 unpack(int packed, int sij) {
				return unpack(packed, sij, sij);
			}

			constexpr static int_vec2 unpack(int packed, int si, int sj) {
				uint p = packed;
				uint i = packed / sj;
				uint j = packed - i * sj;
				assert(int_vec2(i, j).in_bounds(si, sj));
				return int_vec2(i, j);
			}

			constexpr bool in_bounds(int sij) const {
				return in_bounds(sij, sij);
			}

			constexpr bool in_bounds(int si, int sj) const {
				return i >= 0 && i < si && j >= 0 && j < sj;
			}

			constexpr int_vec2 operator+() const {
				return {i, j};
			}

			constexpr int_vec2 operator-() const {
				return {-i, -j};
			}

			constexpr friend int_vec2 operator+(int_vec2 a, int_vec2 b) {
				return {T(a.i + b.i), T(a.j + b.j)};
			}

			constexpr friend int_vec2 operator+(T a, int_vec2 b) {
				return {T(a + b.i), T(a + b.j)};
			}

			constexpr friend int_vec2 operator+(int_vec2 a, T b) {
				return {T(a.i + b), T(a.j + b)};
			}

			constexpr friend int_vec2 operator-(int_vec2 a, int_vec2 b) {
				return {T(a.i - b.i), T(a.j - b.j)};
			}

			constexpr friend int_vec2 operator-(T a, int_vec2 b) {
				return {T(a - b.i), T(a - b.j)};
			}

			constexpr friend int_vec2 operator-(int_vec2 a, T b) {
				return {T(a.i - b), T(a.j - b)};
			}

			constexpr friend int_vec2 operator*(int_vec2 a, int_vec2 b) {
				return {T(a.i * b.i), T(a.j * b.j)};
			}

			constexpr friend int_vec2 operator*(T a, int_vec2 b) {
				return {T(a * b.i), T(a * b.j)};
			}

			constexpr friend int_vec2 operator*(int_vec2 a, T b) {
				return {T(a.i * b), T(a.j * b)};
			}

			constexpr friend int_vec2 operator/(int_vec2 a, int_vec2 b) {
				return {T(a.i / b.i), T(a.j / b.j)};
			}

			constexpr friend int_vec2 operator/(T a, int_vec2 b) {
				return {T(a / b.i), T(a / b.j)};
			}

			constexpr friend int_vec2 operator/(int_vec2 a, T b) {
				return {T(a.i / b), T(a.j / b)};
			}

			constexpr friend int_vec2 operator%(int_vec2 a, int_vec2 b) {
				return {T(a.i % b.i), T(a.j % b.j)};
			}

			constexpr friend int_vec2 operator%(T a, int_vec2 b) {
				return {T(a % b.i), T(a % b.j)};
			}

			constexpr friend int_vec2 operator%(int_vec2 a, T b) {
				return {T(a.i % b), T(a.j % b)};
			}

			constexpr int_vec2 operator+=(int_vec2 a) {
				i += a.i;
				j += a.j;
				return *this;
			}

			constexpr int_vec2 operator+=(T a) {
				i += a;
				j += a;
				return *this;
			}

			constexpr int_vec2 operator-=(int_vec2 a) {
				i -= a.i;
				j -= a.j;
				return *this;
			}

			constexpr int_vec2 operator-=(T a) {
				i -= a;
				j -= a;
				return *this;
			}

			constexpr int_vec2 operator*=(int_vec2 a) {
				i *= a.i;
				j *= a.j;
				return *this;
			}

			constexpr int_vec2 operator*=(T a) {
				i *= a;
				j *= a;
				return *this;
			}

			constexpr int_vec2 operator/=(int_vec2 a) {
				i /= a.i;
				j /= a.j;
				return *this;
			}

			constexpr int_vec2 operator/=(T a) {
				i /= a;
				j /= a;
				return *this;
			}

			constexpr int_vec2 operator%=(int_vec2 a) {
				i %= a.i;
				j %= a.j;
				return *this;
			}

			constexpr int_vec2 operator%=(T a) {
				i %= a;
				j %= a;
				return *this;
			}

			constexpr friend bool operator==(int_vec2 a, int_vec2 b) {
				return a.i == b.i && a.j == b.j;
			}

			constexpr friend bool operator!=(int_vec2 a, int_vec2 b) {
				return a.i != b.i || a.j != b.j;
			}

			constexpr friend bool operator<(int_vec2 a, int_vec2 b) {
				return a.i < b.i || a.i == b.i && a.j < b.j;
			}

			constexpr friend bool operator<=(int_vec2 a, int_vec2 b) {
				return a.i <= b.i || a.i == b.i && a.j <= b.j;
			}

			constexpr friend bool operator>(int_vec2 a, int_vec2 b) {
				return a.i > b.i || a.i == b.i && a.j > b.j;
			}

			constexpr friend bool operator>=(int_vec2 a, int_vec2 b) {
				return a.i >= b.i || a.i == b.i && a.j >= b.j;
			}

			friend ostream& operator<<(ostream& out, int_vec2 a) {
				out << "(" << (int) a.i << ", " << (int) a.j << ")";
				return out;
			}
		};

		template <class T>
		requires(same_as<T, char> || same_as<T, short> || same_as<T, int>)
		struct int_vec3 {
			T i;
			T j;
			T k;

			template <class U>
			constexpr int_vec3(int_vec3<U> a) : i(a.i), j(a.j), k(a.k) {
				assert(i == a.i);
				assert(j == a.j);
				assert(k == a.k);
			}

			constexpr int_vec3() : i(0), j(0), k(0) {
			}

			constexpr int_vec3(T i, T j, T k) : i(i), j(j), k(k) {
			}

			constexpr static int_vec3 dir(int index) {
				constexpr T DIRS[6][3] = {
					{-1, 0, 0},
					{1, 0, 0},
					{0, -1, 0},
					{0, 1, 0},
					{0, 0, -1},
					{0, 0, 1},
				};
				return {DIRS[index][0], DIRS[index][1], DIRS[index][2]};
			}

			constexpr int dir_index() const {
				assert((i != 0) + (j != 0) + (k != 0) == 1);
				return i < 0 ? 0 : i > 0 ? 1 : j < 0 ? 2 : j > 0 ? 3 : k < 0 ? 4 : 5;
			}

			constexpr int_vec3 min(int_vec3 a) const {
				return {std::min(i, a.i), std::min(j, a.j), std::min(k, a.k)};
			}

			constexpr int_vec3 max(int_vec3 a) const {
				return {std::max(i, a.i), std::max(j, a.j), std::max(k, a.k)};
			}

			constexpr int_vec3 clamp(int_vec3 min, int_vec3 max) const {
				return {
					std::clamp(i, min.i, max.i), std::clamp(j, min.j, max.j), std::clamp(k, min.k, max.k)};
			}

			int_vec3 abs() const {
				return {std::abs(i), std::abs(j), std::abs(k)};
			}

			int norm() const {
				return std::abs(i) + std::abs(j) + std::abs(k);
			}

			constexpr int dot(int_vec3 a) const {
				return i * a.i + j * a.j + k * a.k;
			}

			constexpr int pack(int sijk) const {
				return pack(sijk, sijk, sijk);
			}

			constexpr int pack(int si, int sj, int sk) const {
				assert(in_bounds(si, sj, sk));
				return (i * sj + j) * sk + k;
			}

			constexpr int pack_if_in_bounds(int sijk) const {
				return pack_if_in_bounds(sijk, sijk, sijk);
			}

			constexpr int pack_if_in_bounds(int si, int sj, int sk) const {
				if (!in_bounds(si, sj, sk))
					return -1;
				return (i * sj + j) * sk + k;
			}

			constexpr static int_vec3 unpack(int packed, int sijk) {
				return unpack(packed, sijk, sijk, sijk);
			}

			constexpr static int_vec3 unpack(int packed, int si, int sj, int sk) {
				uint p = packed;
				uint ij = p / sk;
				uint k = p - ij * sk;
				uint i = ij / sj;
				uint j = ij - i * sj;
				assert(int_vec3(i, j, k).in_bounds(si, sj, sk));
				return int_vec3(i, j, k);
			}

			constexpr bool in_bounds(int sijk) const {
				return in_bounds(sijk, sijk, sijk);
			}

			constexpr bool in_bounds(int si, int sj, int sk) const {
				return i >= 0 && i < si && j >= 0 && j < sj && k >= 0 && k < sk;
			}

			constexpr int_vec3 operator+() const {
				return {i, j, k};
			}

			constexpr int_vec3 operator-() const {
				return {-i, -j, -k};
			}

			constexpr friend int_vec3 operator+(int_vec3 a, int_vec3 b) {
				return {a.i + b.i, a.j + b.j, a.k + b.k};
			}

			constexpr friend int_vec3 operator+(T a, int_vec3 b) {
				return {a + b.i, a + b.j, a + b.k};
			}

			constexpr friend int_vec3 operator+(int_vec3 a, T b) {
				return {a.i + b, a.j + b, a.k + b};
			}

			constexpr friend int_vec3 operator-(int_vec3 a, int_vec3 b) {
				return {a.i - b.i, a.j - b.j, a.k - b.k};
			}

			constexpr friend int_vec3 operator-(T a, int_vec3 b) {
				return {a - b.i, a - b.j, a - b.k};
			}

			constexpr friend int_vec3 operator-(int_vec3 a, T b) {
				return {a.i - b, a.j - b, a.k - b};
			}

			constexpr friend int_vec3 operator*(int_vec3 a, int_vec3 b) {
				return {a.i * b.i, a.j * b.j, a.k * b.k};
			}

			constexpr friend int_vec3 operator*(T a, int_vec3 b) {
				return {a * b.i, a * b.j, a * b.k};
			}

			constexpr friend int_vec3 operator*(int_vec3 a, T b) {
				return {a.i * b, a.j * b, a.k * b};
			}

			constexpr friend int_vec3 operator/(int_vec3 a, int_vec3 b) {
				return {a.i / b.i, a.j / b.j, a.k / b.k};
			}

			constexpr friend int_vec3 operator/(T a, int_vec3 b) {
				return {a / b.i, a / b.j, a / b.k};
			}

			constexpr friend int_vec3 operator/(int_vec3 a, T b) {
				return {a.i / b, a.j / b, a.k / b};
			}

			constexpr friend int_vec3 operator%(int_vec3 a, int_vec3 b) {
				return {a.i % b.i, a.j % b.j, a.k % b.k};
			}

			constexpr friend int_vec3 operator%(T a, int_vec3 b) {
				return {a % b.i, a % b.j, a % b.k};
			}

			constexpr friend int_vec3 operator%(int_vec3 a, T b) {
				return {a.i % b, a.j % b, a.k % b};
			}

			constexpr int_vec3 operator+=(int_vec3 a) {
				i += a.i;
				j += a.j;
				k += a.k;
				return *this;
			}

			constexpr int_vec3 operator+=(T a) {
				i += a;
				j += a;
				k += a;
				return *this;
			}

			constexpr int_vec3 operator-=(int_vec3 a) {
				i -= a.i;
				j -= a.j;
				k -= a.k;
				return *this;
			}

			constexpr int_vec3 operator-=(T a) {
				i -= a;
				j -= a;
				k -= a;
				return *this;
			}

			constexpr int_vec3 operator*=(int_vec3 a) {
				i *= a.i;
				j *= a.j;
				k *= a.k;
				return *this;
			}

			constexpr int_vec3 operator*=(T a) {
				i *= a;
				j *= a;
				k *= a;
				return *this;
			}

			constexpr int_vec3 operator/=(int_vec3 a) {
				i /= a.i;
				j /= a.j;
				k /= a.k;
				return *this;
			}

			constexpr int_vec3 operator/=(T a) {
				i /= a;
				j /= a;
				k /= a;
				return *this;
			}

			constexpr int_vec3 operator%=(int_vec3 a) {
				i %= a.i;
				j %= a.j;
				k %= a.k;
				return *this;
			}

			constexpr int_vec3 operator%=(T a) {
				i %= a;
				j %= a;
				k %= a;
				return *this;
			}

			constexpr friend bool operator==(int_vec3 a, int_vec3 b) {
				return a.i == b.i && a.j == b.j && a.k == b.k;
			}

			constexpr friend bool operator!=(int_vec3 a, int_vec3 b) {
				return a.i != b.i || a.j != b.j || a.k != b.k;
			}

			constexpr friend bool operator<(int_vec3 a, int_vec3 b) {
				return a.i < b.i || a.i == b.i && (a.j < b.j || a.j == b.j && a.k < b.k);
			}

			constexpr friend bool operator<=(int_vec3 a, int_vec3 b) {
				return a.i <= b.i || a.i == b.i && (a.j <= b.j || a.j == b.j && a.k <= b.k);
			}

			constexpr friend bool operator>(int_vec3 a, int_vec3 b) {
				return a.i > b.i || a.i == b.i && (a.j > b.j || a.j == b.j && a.k > b.k);
			}

			constexpr friend bool operator>=(int_vec3 a, int_vec3 b) {
				return a.i >= b.i || a.i == b.i && (a.j >= b.j || a.j == b.j && a.k >= b.k);
			}

			friend ostream& operator<<(ostream& out, int_vec3 a) {
				out << "(" << (int) a.i << ", " << (int) a.j << ", " << (int) a.k << ")";
				return out;
			}
		};

		using cvec2 = int_vec2<char>;
		using svec2 = int_vec2<short>;
		using ivec2 = int_vec2<int>;
		using cvec3 = int_vec3<char>;
		using svec3 = int_vec3<short>;
		using ivec3 = int_vec3<int>;
	} // namespace ds
	using namespace ds;

	namespace beam_search {
		// (state) -> score
		template <class T, class State, class Score>
		concept get_score =
			totally_ordered<Score> && invocable<T, State&> && same_as<invoke_result_t<T, State&>, Score>;

		// (state, move) -> void
		template <class T, class State, class MoveId>
		concept apply_move =
			invocable<T, State&, MoveId> && same_as<invoke_result_t<T, State&, MoveId>, void>;

		// (state) -> void
		template <class T, class State>
		concept undo_move = invocable<T, State&> && same_as<invoke_result_t<T, State&>, void>;

		// (state) -> void
		// see also: add_candidate
		template <class T, class State>
		concept enumerate_candidates = invocable<T, State&> && same_as<invoke_result_t<T, State&>, void>;

		// (turn) -> void
		// see also: candidates_to_filter
		template <class T>
		concept filter_candidates = invocable<T, int> && same_as<invoke_result_t<T, int>, void>;

		template <class State, totally_ordered Score, class MoveId, MoveId UnusedMoveId, class CandidateData,
			class Direction = greater<Score>, int HashBucketBits = 20>
		class beam_search {
		private:
			struct candidate {
				int index; // index in orig_candidates
				int parent;
				MoveId move_id;
				CandidateData data;
				ull hash;
			};
			struct orig_candidate {
				int parent;
				MoveId move_id;
				bool chosen;
			};

			Direction dir = {};
			int current_parent = 0;
			hash_imap<int, HashBucketBits> best_indices;
			bool enumerating = false;
			bool filtering = false;
			vector<candidate> candidates;
			vector<orig_candidate> orig_candidates;

			void clear_candidates() {
				candidates.clear();
				orig_candidates.clear();
			}

		public:
			Score best_score = 0;
			int max_turn = -1;

			beam_search() {
			}

			beam_search(Direction dir) : dir(dir) {
			}

			void add_candidate(MoveId move_id, CandidateData data, ull hash) {
				assert(msg(enumerating, "not enumerating now"));
				candidates.emplace_back((int) candidates.size(), current_parent, move_id, data, hash);
				orig_candidates.emplace_back(current_parent, move_id);
			}

			vector<candidate>& candidates_to_filter() {
				assert(msg(filtering, "not filtering now"));
				return candidates;
			}

			// CAUTION: not stable
			template <predicate<candidate&, candidate&> CandidateDirection>
			void remove_duplicates(CandidateDirection candidate_direction) {
				assert(msg(filtering, "not filtering now"));
				best_indices.clear();
				int n = (int) candidates.size();
				for (int i = 0; i < n; i++) {
					candidate& cand = candidates[i];
					if (best_indices.access(cand.hash)) {
						int j = best_indices.get();
						candidate& cand2 = candidates[j];
						if (candidate_direction(cand, cand2)) {
							swap(candidates[i], candidates[j]);
						}
						swap(candidates[i], candidates[--n]);
						i--;
					} else {
						best_indices.set(i);
					}
				}
				candidates.resize(n);
			}

			template <get_score<State, Score> GetScore, apply_move<State, MoveId> ApplyMove,
				enumerate_candidates<State> EnumerateCandidates, filter_candidates FilterCandidates>
			vector<MoveId> run(const State& initial_state, GetScore get_score, ApplyMove apply_move,
				EnumerateCandidates enumerate_candidates, FilterCandidates filter_candidates) {
				struct node {
					State state;
					int history_index;
				};
				struct history {
					MoveId move_id;
					int parent;
				};
				vector<node> src;
				vector<node> dst;
				increasing_vector<history> hs;
				int turn = 0;

				// set initial state
				src.emplace_back(initial_state, -1);

				while (true) {
					int num_states = (int) src.size();

					clear_candidates();
					if (max_turn == -1 || turn < max_turn) {
						// enumerate candidates
						enumerating = true;
						for (int i = 0; i < num_states; i++) {
							current_parent = i;
							enumerate_candidates(src[i].state);
						}
						enumerating = false;

						// filer candiadtes
						filtering = true;
						filter_candidates(turn);
						filtering = false;
					}

					// check if finished
					if (candidates.empty()) {
						assert(msg(num_states > 0, "no states at the end"));

						// pick the best state
						best_score = get_score(src[0].state);
						int best_index = 0;
						for (int i = 1; i < num_states; i++) {
							Score score = get_score(src[i].state);
							if (dir(score, best_score)) {
								best_score = score;
								best_index = i;
							}
						}

						// restore moves
						vector<MoveId> res;
						int history_top = src[best_index].history_index;
						while (history_top != -1) {
							history& h = hs[history_top];
							res.push_back(h.move_id);
							history_top = h.parent;
						}
						reverse(res.begin(), res.end());
						return res;
					}

					// compute next states
					dst.clear();
					for (const auto& cand : candidates) {
						const auto& src_node = src[cand.parent];
						dst.emplace_back(src_node.state, hs.size());
						apply_move(dst.back().state, cand.move_id);
						hs.emplace_back(cand.move_id, src_node.history_index);
					}
					src.swap(dst);
					turn++;
				}
			}

			template <get_score<State, Score> GetScore, apply_move<State, MoveId> ApplyMove,
				undo_move<State> UndoMove, enumerate_candidates<State> EnumerateCandidates,
				filter_candidates FilterCandidates>
			vector<MoveId> run_tree(const State& initial_state, GetScore get_score, ApplyMove apply_move,
				UndoMove undo_move, EnumerateCandidates enumerate_candidates,
				FilterCandidates filter_candidates) {
				constexpr MoveId UNDO = UnusedMoveId;
				struct tour {
					vector<MoveId> src;
					vector<MoveId> dst;

					void move(const MoveId& move_id) {
						dst.push_back(move_id);
					}

					int position() {
						return (int) dst.size();
					}

					void swap() {
						src.swap(dst);
						dst.clear();
					}
				} tour;
				vector<MoveId> global_path;
				vector<MoveId> path;
				vector<orig_candidate> leaves;
				State st = initial_state;
				int turn = 0;
				int level = 0;
				int next_start_pos = 0;

				auto global_move = [&](const MoveId& move_id) {
					apply_move(st, move_id);
					global_path.push_back(move_id);
					level++;
				};

				auto global_undo = [&]() {
					undo_move(st);
					global_path.pop_back();
					level--;
				};

				while (true) {
					bool has_next_turn = max_turn == -1 || turn < max_turn;

					// compute the next tour
					int pos = next_start_pos;
					int prev_root_level = level;
					int next_root_level = numeric_limits<int>::max();
					orig_candidate best_leaf = {-1, MoveId{}, false};
					enumerating = true;
					clear_candidates();
					if (turn == 0) {
						best_score = get_score(st);
						best_leaf.chosen = true;
						if (has_next_turn) {
							current_parent = tour.position();
							enumerate_candidates(st);
						}
					} else {
						for (const orig_candidate& leaf : leaves) {
							int parent_pos = leaf.parent;

							// visit the parent of the leaf node
							if (pos < parent_pos) {
								// visit the LCA
								path.clear();
								do {
									auto move = tour.src[pos++];
									if (move == UNDO) {
										if (path.empty()) {
											global_undo();
											tour.move(UNDO);
											next_root_level = min(next_root_level, level);
										} else {
											path.pop_back();
										}
									} else {
										path.push_back(move);
									}
								} while (pos < parent_pos);

								// go directly to the parent
								for (auto move : path) {
									global_move(move);
									tour.move(move);
								}
							} // now we are at the parent of the leaf node

							// visit the leaf node
							apply_move(st, leaf.move_id);
							tour.move(leaf.move_id);

							Score score = get_score(st);
							if (!best_leaf.chosen || dir(score, best_score)) {
								best_score = score;
								best_leaf = leaf;
							}
							if (has_next_turn) {
								current_parent = tour.position();
								enumerate_candidates(st);
							}

							// leave the leaf node
							undo_move(st);
							tour.move(UNDO);
						}
					}
					next_root_level = min(next_root_level, level);
					enumerating = false;

					filtering = true;
					filter_candidates(turn);
					filtering = false;

					if (candidates.empty()) {
						assert(best_leaf.chosen);
						// undo to the root level
						while (level > prev_root_level) {
							global_undo();
						}
						// visit the best leaf
						pos = next_start_pos;
						while (pos < best_leaf.parent) {
							auto move = tour.src[pos++];
							if (move == UNDO) {
								global_undo();
							} else {
								global_move(move);
							}
						}
						if (best_leaf.parent != -1) {
							global_move(best_leaf.move_id);
						}
						return global_path;
					}

					// finalize the next tour
					tour.swap();
					turn++;

					// collect the next leaf nodes, in the original order
					leaves.clear();
					for (const candidate& cand : candidates) {
						orig_candidates[cand.index].chosen = true;
					}
					for (const orig_candidate& cand : orig_candidates) {
						if (!cand.chosen)
							continue;
						leaves.push_back(cand);
					}

					// undo to the next root level
					while (level > next_root_level) {
						global_undo();
					}

					// adjust the next starting position
					next_start_pos = next_root_level - prev_root_level;
				}
			}
		};

		class beam_width_manager {
		private:
			double prev_time = 0;
			double moving_average_time = 0;
			vector<double> progress_history;
			vector<double> time_history;
			vector<int> width_history;
			int last_width = 0;
			int count = 0;

		public:
			int window_size = 50;
			int default_width;

			beam_width_manager(int default_width) : default_width(default_width) {
			}

			int next(double progress, double time, double time_limit) {
				progress_history.push_back(progress);
				time_history.push_back(time);
				width_history.push_back(last_width);
				count++;
				if (count <= window_size) {
					return last_width = default_width;
				}
				int i1 = count - 1 - window_size;
				int i2 = count - 1;
				double progress_sum = progress_history[i2] - progress_history[i1];
				double time_sum = time_history[i2] - time_history[i1];
				if (progress_sum == 0 || time_sum == 0) {
					// window size is too small
					window_size *= 2;
					return last_width = default_width;
				}
				int width_sum = 0;
				for (int i = i1 + 1; i <= i2; i++) {
					width_sum += width_history[i];
				}
				double progress_per_turn = progress_sum / window_size;
				double time_per_width = time_sum / width_sum;
				double left_time = time_limit - time;
				double left_progress = 1 - progress;
				if (left_time <= 0 || left_progress <= 0)
					return 1;
				double left_turn = left_progress / progress_per_turn;
				double left_time_per_turn = left_time / left_turn;
				double left_width_per_turn = left_time_per_turn / time_per_width;
				return last_width = (int) round(left_width_per_turn);
			}

			void report(int actual_last_width) {
				last_width = actual_last_width;
			}
		};
	} // namespace beam_search

	namespace simulated_annealing {
		// (state) -> score
		template <class T, class State, class Score>
		concept get_score =
			totally_ordered<Score> && invocable<T, State&> && same_as<invoke_result_t<T, State&>, Score>;

		// (iter) -> progress
		template <class T>
		concept update_progress = invocable<T, int> && same_as<invoke_result_t<T, int>, double>;

		// (state, tolerance) -> accepted
		template <class T, class State>
		concept try_transition =
			invocable<T, State&, double> && same_as<invoke_result_t<T, State&, double>, bool>;

		template <class State, totally_ordered Score, class Direction = greater<Score>>
		class simulated_annealing {
		private:
			Direction dir = {};

		public:
			int clock_interval = 10;
			double t_from = 100;
			double t_to = 0.01;
			double progress = 0;
			int num_iterations = 0;
			int num_acceptances = 0;
			int num_rejections = 0;
			bool use_linear_temp = false;
			Score best_score = 0;

			simulated_annealing() {
			}

			simulated_annealing(Direction dir) : dir(dir) {
			}

			template <get_score<State, Score> GetScore, update_progress UpdateProgress,
				try_transition<State> TryTransition>
			State run(const State& initial_state, rngen& rng, GetScore get_score,
				UpdateProgress update_progress, TryTransition try_transition,
				function<void(State&, Score, int, double)> best_updated = nullptr) {
				State state = initial_state;
				Score score = get_score(state);
				State best_state = state;
				best_score = score;

				num_iterations = 0;
				num_acceptances = 0;
				num_rejections = 0;
				int interval = clock_interval;
				progress = 0;
				double t = t_from;
				while (true) {
					if (--interval <= 0) {
						progress = update_progress(num_iterations);
						if (progress >= 1)
							break;
						t = use_linear_temp ? lerp(t_from, t_to, progress)
											: exp_interp(t_from, t_to, progress);
						interval = clock_interval;
					}
					double tolerance = t * -log(rng.next_float());
					if (try_transition(state, tolerance)) {
						num_acceptances++;
						score = get_score(state);
						if (dir(score, best_score)) {
							best_state = state;
							best_score = score;
							if (best_updated) {
								best_updated(state, score, num_iterations, t);
							}
						}
					} else {
						num_rejections++;
					}
					num_iterations++;
				}
				return best_state;
			}
		};
	} // namespace simulated_annealing

	namespace dijkstra {
		// (vertex) -> index
		template <class T, class Vertex>
		concept get_index = invocable<T, Vertex> && same_as<invoke_result_t<T, Vertex>, int>;

		// (vertex) -> is_goal
		template <class T, class Vertex>
		concept is_goal = invocable<T, Vertex> && same_as<invoke_result_t<T, Vertex>, bool>;

		// (vertex, distance) -> void
		template <class T, class Vertex, class Weight>
		concept visit_adjacent_vertices =
			invocable<T, Vertex, Weight> && same_as<invoke_result_t<T, Vertex, Weight>, void>;

		template <class Vertex, class Weight, Weight Infinity, int MaxVertexIndex>
		requires(integral<Weight> || floating_point<Weight>)
		class dijkstra {
		private:
			using vw = pair<Vertex, Weight>;
			static constexpr int VERTEX_ARRAY_SIZE = MaxVertexIndex + 1;
			vector<vw> toVisit;
			bool visiting = false;

		public:
			array<bool, VERTEX_ARRAY_SIZE> visited;
			array<Weight, VERTEX_ARRAY_SIZE> distance;
			array<optional<Vertex>, VERTEX_ARRAY_SIZE> previous;

			dijkstra() {
			}

			// - get_index: `(vertex) -> index`
			// - is_goal: `(vertex) -> bool`
			// - visit_adjacent_vertices: `(vertex, distance) -> void`
			template <get_index<Vertex> GetIndex, is_goal<Vertex> IsGoal,
				visit_adjacent_vertices<Vertex, Weight> VisitAdjacentVertices>
			void run(const vector<Vertex>& starts, GetIndex get_index, IsGoal is_goal,
				VisitAdjacentVertices visit_adjacent_vertices) {
				auto comp = [](vw& a, vw& b) {
					return a.second > b.second;
				};

				visited.fill(false);
				previous.fill(nullopt);
				distance.fill(Infinity);
				priority_queue<vw, vector<vw>, decltype(comp)> q(comp);
				for (auto& st : starts) {
					distance[get_index(st)] = Weight(0);
					q.emplace(st, Weight(0));
				}

				while (!q.empty()) {
					auto [from, dist] = q.top();
					q.pop();
					int fromi = get_index(from);
					if (visited[fromi])
						continue;
					visited[fromi] = true;
					if (is_goal(from)) {
						return;
					}

					visiting = true;
					toVisit.clear();
					visit_adjacent_vertices(from, dist);
					visiting = false;

					for (vw& pair : toVisit) {
						Vertex to = pair.first;
						int toi = get_index(to);
						Weight new_dist = pair.second;
						if (new_dist < distance[toi]) {
							distance[toi] = new_dist;
							previous[toi] = from;
							q.emplace(to, new_dist);
						}
					}
				}
			}

			void visit(Vertex vertex, Weight total_distance) {
				assert(visiting);
				toVisit.emplace_back(vertex, total_distance);
			}

			template <get_index<Vertex> GetIndex>
			vector<Vertex> restore_path(Vertex goal, GetIndex get_index) {
				assert(msg("goal is not reached", visited[get_index(goal)]));
				vector<Vertex> res;
				Vertex v = goal;
				while (true) {
					res.push_back(v);
					if (!previous[get_index(v)])
						break;
					v = previous[get_index(v)].value();
				}
				reverse(res.begin(), res.end());
				return res;
			}
		};
	}; // namespace dijkstra

	namespace file {
		string pad4(int n) {
			return (n < 0 || n >= 1000 ? "" : n < 10 ? "000" : n < 100 ? "00" : "0") + tos(n);
		}

		string input_file_name(int seed) {
			return "in/" + pad4(seed) + ".txt";
		}

		string output_file_name(int seed) {
			return "out/" + pad4(seed) + ".txt";
		}

		string movie_file_name(int seed) {
			return "mov/" + pad4(seed) + ".smv";
		}

		bool write_text(string fileName, string text, bool append = false) {
			ofstream fout;
			fout.open(fileName, append ? ios::out | ios::app : ios::out);
			if (!fout)
				return false;
			fout << text;
			fout.close();
			return true;
		}

		pair<string, bool> read_text(string fileName) {
			ifstream fin;
			fin.open(fileName, ios::in);
			if (!fin)
				return make_pair("", false);
			string res;
			string line;
			while (getline(fin, line)) {
				res += line;
				res += "\n";
			}
			return make_pair(res, true);
		}

		void write_text_assert(string fileName, string text, bool append = false) {
			auto res = write_text(fileName, text, append);
			assert(res);
		}

		string read_text_assert(string fileName) {
			auto res = read_text(fileName);
			assert(res.second);
			return res.first;
		}
	} // namespace file
} // namespace shr
using namespace shr::basic;
using namespace shr::ds;
using namespace shr::beam_search;
using namespace shr::simulated_annealing;
using namespace shr::dijkstra;
using namespace shr::random;
using namespace shr::timer;
using namespace shr::tracer;
using namespace shr::file;

//
// --- macros ---
//

#define rep(i, from, until) for (int i = (from); i < (until); i++)
#define repr(i, from, until) for (int i = (until) - 1; i >= (from); i--)
#define rep0(i, until) rep(i, 0, until)
#define rep0r(i, until) repr(i, 0, until)

//
// --- movie lib ---
//

class movie {
private:
	bool is_open = false;
	string file_name;
	ofstream out;

#ifdef ONLINE_JUDGE
	void write_instruction(const string& it, const vector<double>& argf, const vector<string>& args) {
	}
#else
	void write_instruction(const string& it, const vector<double>& argf, const vector<string>& args) {
		assert(msg(file_name != "", "file name is not set"));
		if (!is_open) {
			is_open = true;
			out = ofstream(file_name, ios_base::out | ios_base::binary);
			out.write("smv", 3);
		}
		write_string(it);
		for (double f : argf) {
			write_float(f);
		}
		for (auto& s : args) {
			write_string(s);
		}
		if (it == "n") {
			out.flush(); // flush at the end of each frame
		}
	}

	void write_float(double a) {
		float f = (float) a;
		out.write((char*) &f, 4);
	}

	void write_string(const string& str) {
		out.write(str.c_str(), str.length() + 1);
	}
#endif

public:
	static constexpr int LEFT = 0;
	static constexpr int CENTER = 1;
	static constexpr int RIGHT = 2;

	movie() {
	}

	void set_file(string file) {
		assert(!is_open);
		file_name = file;
	}

	void close_file() {
		if (!is_open)
			return;
		is_open = false;
		out.close();
	}

	void fill(double rgb, double a = 1.0) {
		fill(rgb, rgb, rgb, a);
	}

	void fill(double r, double g, double b, double a = 1.0) {
		write_instruction("f", {r, g, b, a}, {});
	}

	void stroke(double rgb, double a = 1.0) {
		stroke(rgb, rgb, rgb, a);
	}

	void stroke(double r, double g, double b, double a = 1.0) {
		write_instruction("s", {r, g, b, a}, {});
	}

	void no_fill() {
		write_instruction("nf", {}, {});
	}

	void no_stroke() {
		write_instruction("ns", {}, {});
	}

	void comment(string text) {
		write_instruction("cm", {}, {text});
	}

	void tooltip(string text) {
		write_instruction("tt", {}, {text});
	}

	void no_tooltip() {
		write_instruction("nt", {}, {});
	}

	template <class T>
	void with_tooltip(string text, T f) {
		tooltip(text);
		f();
		no_tooltip();
	}

	void stroke_weight(double weight) {
		write_instruction("sw", {weight}, {});
	}

	void text_size(double size) {
		write_instruction("ts", {size}, {});
	}

	void text_align(int align) {
		write_instruction("ta", {(double) align}, {});
	}

	void rect(double x, double y, double w, double h) {
		write_instruction("r", {x, y, w, h}, {});
	}

	void circle(double x, double y, double r) {
		write_instruction("c", {x, y, r}, {});
	}

	void ellipse(double x, double y, double rx, double ry) {
		write_instruction("e", {x, y, rx, ry}, {});
	}

	void line(double x1, double y1, double x2, double y2) {
		write_instruction("l", {x1, y1, x2, y2}, {});
	}

	void text(string str, double x, double y) {
		write_instruction("t", {x, y}, {str});
	}

	void transform(double e00, double e01, double e10, double e11) {
		write_instruction("tf", {e00, e01, e10, e11}, {});
	}

	void translate(double tx, double ty) {
		write_instruction("tr", {tx, ty}, {});
	}

	void rotate(double ang) {
		write_instruction("ro", {ang}, {});
	}

	void scale(double s) {
		scale(s, s);
	}

	void scale(double sx, double sy) {
		write_instruction("sc", {sx, sy}, {});
	}

	void push() {
		write_instruction("pu", {}, {});
	}

	void pop() {
		write_instruction("po", {}, {});
	}

	void end_frame() {
		write_instruction("n", {}, {});
	}

	void target(string name) {
		write_instruction("tg", {}, {name});
	}
};

movie mov;

// --------------------------------------------------
// main part
// --------------------------------------------------

bool isLocal = false;
bool render = false;

// define N and stuff here
constexpr int INF = 1'000'000'000;
constexpr int N = 50;
constexpr int N2 = N * N;
constexpr int MAX_M = 1600;
constexpr int T = 800;
constexpr int MAX_STATIONS = 300;
#define repn(i) rep0(i, N)
#define rep2() rep0(i, N) rep0(j, N)
#define repn2(i) rep0(i, N2)
#define rept(i) rep0(i, T)
#define repm(i) rep0(i, m)

constexpr int DIR_U = 0;
constexpr int DIR_D = 1;
constexpr int DIR_L = 2;
constexpr int DIR_R = 3;

constexpr int STATION_COST = 5000;
constexpr int RAIL_COST = 100;

struct Problem {
	int m;
	int k;
	vector<array<ivec2, 2>> poss;
	vector<int> pairDists;
	array<vector<int>, N2> stationToPlaces = {};

	void load(istream& in) {
		int n, t;
		in >> n >> m >> k >> t;
		assert(n == N && t == T);
		assert(50 <= m && m <= 1600);
		assert(11000 <= k && k <= 20000);
		repm(i) {
			ivec2 home;
			ivec2 work;
			in >> home.i >> home.j >> work.i >> work.j;
			poss.push_back({home, work});
			pairDists.push_back((home - work).mnorm());
		}

		// each station can cover places within 2 manhattan distance
		rep2() {
			ivec2 spos = {i, j};
			int sposi = spos.pack(N);
			rep0(place, m * 2) {
				ivec2 ppos = poss[place / 2][place % 2];
				if ((ppos - spos).mnorm() <= 2) {
					stationToPlaces[sposi].push_back(place);
				}
			}
		}
	}
};

class Tile {
private:
	char data = 0;

	Tile(char data) : data(data) {
	}

public:
	Tile() {
	}

	static Tile empty() {
		return Tile(0);
	}

	// `...`
	// `@@@`
	// `...`
	static Tile lr() {
		return Tile(1);
	}

	// `.@.`
	// `.@.`
	// `.@.`
	static Tile ud() {
		return Tile(2);
	}

	// `...`
	// `@@.`
	// `.@.`
	static Tile ld() {
		return Tile(3);
	}

	// `.@.`
	// `@@.`
	// `...`
	static Tile lu() {
		return Tile(4);
	}

	// `.@.`
	// `.@@`
	// `...`
	static Tile ur() {
		return Tile(5);
	}

	// `...`
	// `.@@`
	// `.@.`
	static Tile rd() {
		return Tile(6);
	}

	// `.@.`
	// `@@@`
	// `.@.`
	static Tile station() {
		return Tile(7);
	}

	static Tile computeTile(ivec2 prev, ivec2 cur, ivec2 next) {
		int dir1 = (prev - cur).dir_index();
		int dir2 = (next - cur).dir_index();
		// U D L R
		constexpr int matrix[4][4] = {
			{-1, 2, 4, 5},
			{2, -1, 3, 6},
			{4, 3, -1, 1},
			{5, 6, 1, -1},
		};
		int kind = matrix[dir1][dir2];
		assert(kind != -1);
		return Tile(kind);
	}

	int kind() const {
		return data;
	}

	int connections() const {
		constexpr int connections[] = {
			0b0000,
			0b1100,
			0b0011,
			0b0110,
			0b0101,
			0b1001,
			0b1010,
			0b1111,
		};
		return connections[data];
	}

	bool connects(int dir) const {
		return (connections() >> dir) & 1;
	}

	bool operator==(const Tile& other) const {
		return data == other.data;
	}

	bool operator!=(const Tile& other) const {
		return data != other.data;
	}
};

struct Connection {
	short index = 0;
	char possibleDirBits = 0;

	bool dirFixed() const {
		return popcount((uint) possibleDirBits) == 1;
	}

	bool hasDir(int dir) const {
		return possibleDirBits & (1 << dir);
	}

	int fixedDir() const {
		assert(dirFixed());
		return countr_zero((uint) possibleDirBits);
	}
};

struct Station {
	using Stations = fast_vector<Station, MAX_STATIONS>;
	cvec2 pos;
	short index;
	fast_vector<Connection, 4> connections;

	Station(ivec2 pos = {-1, -1}, int index = -1) : pos(pos), index(index) {
	}

	bool dirAvailable(int dir) const {
		if (len(connections) == 4)
			return false;
		for (const auto& c : connections) {
			if (c.hasDir(dir) && c.dirFixed())
				return false;
		}
		return true;
	}

	int availableDirBits() const {
		int res = 0;
		rep0(dir, 4) {
			if (dirAvailable(dir))
				res |= 1 << dir;
		}
		return res;
	}

	static int possibleDirBitsForDifference(ivec2 diff) {
		int di = diff.i;
		int dj = diff.j;
		int res = 0b1111;
		if (di >= 0)
			res &= ~(1 << DIR_U);
		if (di <= 0)
			res &= ~(1 << DIR_D);
		if (dj >= 0)
			res &= ~(1 << DIR_L);
		if (dj <= 0)
			res &= ~(1 << DIR_R);
		return res;
	}

	int possibleDirBitsForPos(cvec2 pos2) const {
		return possibleDirBitsForDifference(pos2 - pos) & availableDirBits();
	}

	void addConnectionTo(ivec2 pos, int index) {
		Connection c;
		c.index = index;
		c.possibleDirBits = possibleDirBitsForPos(pos);
		connections.push_back(c);
	}

	static bool connect(Stations& stations, fast_iset<MAX_STATIONS>& updated, int index1, int index2,
		bool allowFail = false) {
		Station& s1 = stations[index1];
		Station& s2 = stations[index2];
		ivec2 diff = s2.pos - s1.pos;
		int dirBits1 = s1.possibleDirBitsForPos(s2.pos);
		int dirBits2 = s2.possibleDirBitsForPos(s1.pos);
		if (allowFail && (dirBits1 == 0 || dirBits2 == 0))
			return false;
		assert(dirBits1);
		assert(dirBits2);
		Connection& c1 = s1.connections.emplace_back();
		c1.index = index2;
		c1.possibleDirBits = dirBits1;
		Connection& c2 = s2.connections.emplace_back();
		c2.index = index1;
		c2.possibleDirBits = dirBits2;
		updated.insert(index1);
		updated.insert(index2);
		if (diff.mnorm() > 1) {
			if (c1.dirFixed() && !s1.restrictOtherDir(stations, updated, c1, allowFail))
				return false;
			if (c2.dirFixed() && !s2.restrictOtherDir(stations, updated, c2, allowFail))
				return false;
		}
		return true;
	}

	// insert 3 between 1 and 2
	static void insert(
		Stations& stations, fast_iset<MAX_STATIONS>& updated, int index1, int index2, int index3) {
		Station& s1 = stations[index1];
		Station& s2 = stations[index2];
		Station& s3 = stations[index3];
		assert(insertable(s1, s2, s3.pos));
		ivec2 diff13 = s3.pos - s1.pos;
		ivec2 diff32 = s2.pos - s3.pos;
		Connection& c1 = s1.getConnectionTo(index2);
		Connection& c2 = s2.getConnectionTo(index1);
		Connection& c31 = s3.connections.emplace_back();
		Connection& c32 = s3.connections.emplace_back();
		c1.possibleDirBits &= Station::possibleDirBitsForDifference(s3.pos - s1.pos);
		c2.possibleDirBits &= Station::possibleDirBitsForDifference(s3.pos - s2.pos);
		assert(c1.possibleDirBits);
		assert(c2.possibleDirBits);
		c1.index = index3;
		c2.index = index3;
		c31.index = index1;
		c32.index = index2;
		c31.possibleDirBits = s3.possibleDirBitsForPos(s1.pos);
		c32.possibleDirBits = s3.possibleDirBitsForPos(s2.pos);
		assert(c31.possibleDirBits);
		assert(c32.possibleDirBits);
		if (c1.dirFixed() && diff13.mnorm() > 1) {
			s1.restrictOtherDir(stations, updated, c1);
		}
		if (c2.dirFixed() && diff32.mnorm() > 1) {
			s2.restrictOtherDir(stations, updated, c2);
		}
		updated.insert(index1);
		updated.insert(index2);
		updated.insert(index3);
	}

	static array<ivec2, 2> connectionAabb(const Station& s1, const Station& s2) {
		ivec2 pos1 = s1.pos;
		ivec2 pos2 = s2.pos;
		auto& c1 = s1.getConnectionTo(s2.index);
		auto& c2 = s2.getConnectionTo(s1.index);
		if (c1.dirFixed()) {
			pos1 += ivec2::dir(c1.fixedDir());
		}
		if (c2.dirFixed()) {
			pos2 += ivec2::dir(c2.fixedDir());
		}
		return {pos1.min(pos2), pos1.max(pos2)};
	}

	static bool insertable(const Station& s1, const Station& s2, ivec2 pos) {
		auto [mins, maxs] = connectionAabb(s1, s2);
		return pos.i >= mins.i && pos.i <= maxs.i && pos.j >= mins.j && pos.j <= maxs.j;
	}

	bool connectable(ivec2 pos) const {
		return possibleDirBitsForPos(pos);
	}

	bool propagate(Stations& stations, fast_iset<MAX_STATIONS>& updated, bool allowFail = false) {
		int fixedDirBits = 0;
		for (auto& c : connections) {
			if (c.dirFixed()) {
				fixedDirBits |= c.possibleDirBits;
			}
		}
		for (auto& c : connections) {
			if (!c.dirFixed()) {
				if (c.possibleDirBits & fixedDirBits) {
					c.possibleDirBits &= ~fixedDirBits;
					assert(c.dirFixed());
					updated.insert(index);
					restrictOtherDir(stations, updated, c);
				}
			}
		}
		rep0(ci1, len(connections)) {
			if (connections[ci1].dirFixed())
				continue;
			auto& c1 = connections[ci1];
			rep0(ci2, len(connections)) {
				auto& c2 = connections[ci2];
				if (ci1 == ci2 || c2.dirFixed())
					continue;
				if (c1.possibleDirBits == c2.possibleDirBits) {
					// fix both directions
					cvec2 p1 = stations[c1.index].pos;
					cvec2 p2 = stations[c2.index].pos;
					cvec2 diff1 = p1 - pos;
					cvec2 diff2 = p2 - pos;
					int cross = diff1.i * diff2.j - diff1.j * diff2.i;
					if (cross == 0) {
						if (allowFail)
							return false;
						trace("WARNING: cross == 0");
					}
					array<int, 2> dirs;
					switch (c1.possibleDirBits) {
					case 0b0101: // LU
						dirs = {DIR_L, DIR_U};
						break;
					case 0b1001: // UR
						dirs = {DIR_U, DIR_R};
						break;
					case 0b1010: // RD
						dirs = {DIR_R, DIR_D};
						break;
					case 0b0110: // DL
						dirs = {DIR_D, DIR_L};
						break;
					default:
						assert(allowFail);
						return false;
					}
					if (cross > 0)
						swap(dirs[0], dirs[1]);
					c1.possibleDirBits = 1 << dirs[0];
					c2.possibleDirBits = 1 << dirs[1];
					updated.insert(index);
					restrictOtherDir(stations, updated, c1);
					restrictOtherDir(stations, updated, c2);
					break;
				}
			}
		}
		return true;
	}

	void collapse(Stations& stations, fast_iset<MAX_STATIONS>& updated, Connection& c, rngen& rng) {
		assert(!c.dirFixed());
		array<int, 2> dirs;
		int dirBits = c.possibleDirBits;
		dirs[0] = countr_zero((uint) dirBits);
		dirBits &= ~(1 << dirs[0]);
		dirs[1] = countr_zero((uint) dirBits);
		c.possibleDirBits = 1 << dirs[rng.next_int(2)];
		updated.insert(index);
		restrictOtherDir(stations, updated, c);
	}

	Connection& getConnectionTo(int index) {
		for (auto& c : connections) {
			if (c.index == index) {
				return c;
			}
		}
		assert(false);
	}

	const Connection& getConnectionTo(int index) const {
		for (const auto& c : connections) {
			if (c.index == index) {
				return c;
			}
		}
		assert(false);
	}

private:
	bool restrictOtherDir(
		Stations& stations, fast_iset<MAX_STATIONS>& updated, Connection& c, bool allowFail = false) const {
		int dir = c.fixedDir();
		int other = c.index;
		Station& os = stations[other];
		Connection& oc = os.getConnectionTo(index);
		bool prevFixed = oc.dirFixed();
		ivec2 diff = os.pos - pos;
		if (dir == DIR_U && diff.i == -1) {
			oc.possibleDirBits &= ~(1 << DIR_D);
		} else if (dir == DIR_D && diff.i == 1) {
			oc.possibleDirBits &= ~(1 << DIR_U);
		} else if (dir == DIR_L && diff.j == -1) {
			oc.possibleDirBits &= ~(1 << DIR_R);
		} else if (dir == DIR_R && diff.j == 1) {
			oc.possibleDirBits &= ~(1 << DIR_L);
		} else {
			return true;
		}
		if (allowFail && oc.possibleDirBits == 0) {
			return false;
		}
		assert(oc.dirFixed());
		if (!prevFixed) {
			updated.insert(other);
		}
		return true;
	}
};

using Stations = Station::Stations;

int orientation(ivec2 a, ivec2 b, ivec2 c) {
	int val = (b.i - a.i) * (c.j - b.j) - (b.j - a.j) * (c.i - b.i);
	return (val > 0) - (val < 0);
}

bool inSegment(ivec2 p, ivec2 a, ivec2 b) {
	ivec2 mins = a.min(b);
	ivec2 maxs = a.max(b);
	return p.i > mins.i && p.i < maxs.i && p.j > mins.j && p.j < maxs.j;
}

bool segmentsIntersect(ivec2 a1, ivec2 b1, ivec2 a2, ivec2 b2) {
	// check by AABB first
	ivec2 mins1 = a1.min(b1);
	ivec2 maxs1 = a1.max(b1);
	ivec2 mins2 = a2.min(b2);
	ivec2 maxs2 = a2.max(b2);
	if (maxs1.i < mins2.i || maxs2.i < mins1.i || maxs1.j < mins2.j || maxs2.j < mins1.j)
		return false;
	if (a1 == a2 || a1 == b2 || b1 == a2 || b1 == b2)
		return false;
	// actual check
	int o1 = orientation(a1, b1, a2);
	int o2 = orientation(a1, b1, b2);
	int o3 = orientation(a2, b2, a1);
	int o4 = orientation(a2, b2, b1);
	if (o1 != o2 && o3 != o4)
		return true;
	if (o1 == 0 && inSegment(a2, a1, b1))
		return true;
	if (o2 == 0 && inSegment(b2, a1, b1))
		return true;
	if (o3 == 0 && inSegment(a1, a2, b2))
		return true;
	if (o4 == 0 && inSegment(b1, a2, b2))
		return true;
	return false;
}

constexpr array<ivec2, 13> COVER_RANGE = {{
	{-1, -1},
	{-1, 0},
	{-1, 1},
	{0, -1},
	{0, 0},
	{0, 1},
	{1, -1},
	{1, 0},
	{1, 1},
	{-2, 0},
	{2, 0},
	{0, -2},
	{0, 2},
}};

constexpr array<cvec2, 13> C_COVER_RANGE = {{
	{-1, -1},
	{-1, 0},
	{-1, 1},
	{0, -1},
	{0, 0},
	{0, 1},
	{1, -1},
	{1, 0},
	{1, 1},
	{-2, 0},
	{2, 0},
	{0, -2},
	{0, 2},
}};

struct StationPlaceInfo {
	cvec2 pos;
	double score = 0;
	char numRailsToBePlaced = 0;
	short index1 = -1;
	short index2 = -1;

	StationPlaceInfo(cvec2 pos, int score, int numRailsToBePlaced, int index1 = -1, int index2 = -1)
		: pos(pos), score(score), numRailsToBePlaced(numRailsToBePlaced), index1(index1), index2(index2) {
	}

	StationPlaceInfo(cvec2 pos, double score, int numRailsToBePlaced, int index1 = -1, int index2 = -1)
		: pos(pos), score(score), numRailsToBePlaced(numRailsToBePlaced), index1(index1), index2(index2) {
	}

	StationPlaceInfo(cvec2 pos) : pos(pos), score(-1), numRailsToBePlaced(0), index1(-1), index2(-1) {
	}
};

struct BsState {
	// should-be-copied data begin
	const Problem* p;
	fast_vector<Station, MAX_STATIONS> stations;
	int money;
	int income;
	int pincome;
	int turn;
	ull hash;
	array<char, N2> dists; // distance to the nearest station
	array<short, N2> dincomes; // delta incomes
	array<short, N2> dpincomes; // delta potential incomes
	bitset<MAX_M * 2> covered; // covered places
	// should-be-copied data end

	bitset<N2> insertable;
	fast_iset<MAX_STATIONS> updated;

	BsState() {
	}

	// copy constructor
	BsState(const BsState& st) {
		this->p = st.p;
		this->stations.copyFrom(st.stations);
		this->money = st.money;
		this->income = st.income;
		this->pincome = st.pincome;
		this->turn = st.turn;
		this->hash = st.hash;
		this->dists = st.dists;
		this->dincomes = st.dincomes;
		this->dpincomes = st.dpincomes;
		this->covered = st.covered;
		updated.clear();
	}

	// assignment operator
	BsState& operator=(const BsState& st) {
		this->p = st.p;
		this->stations.copyFrom(st.stations);
		this->money = st.money;
		this->income = st.income;
		this->pincome = st.pincome;
		this->turn = st.turn;
		this->hash = st.hash;
		this->dists = st.dists;
		this->dincomes = st.dincomes;
		this->dpincomes = st.dpincomes;
		this->covered = st.covered;
		updated.clear();
		return *this;
	}

	static double& powTable(int index) {
		static double table[T + 1] = {};
		return table[index];
	}

	static double& partialIncomeCoeff() {
		static double value = 0;
		return value;
	}

	static ull& hashOfPlace(int place) {
		static ull table[MAX_M * 2] = {};
		return table[place];
	}

	void setup(const Problem& p, rngen& rng) {
		this->p = &p;
		double powBase = 1.001 + p.m / 1600.0 * 0.009;
		rep0(t, T + 1) {
			powTable(t) = pow(powBase, t);
		}
		rep0(i, p.m * 2) {
			hashOfPlace(i) = rng.next_ull();
		}
		partialIncomeCoeff() =
			lerp(0.2, 0.02, linearstep(50, 1600, p.m)) * lerp(0.1, 1, linearstep(11000, 15000, p.k));
		clear();
	}

	void clear() {
		stations.clear();
		money = p->k;
		income = 0;
		pincome = 0;
		turn = 0;
		dists.fill(127);
		dincomes.fill(0);
		dpincomes.fill(0);
		covered.reset();
		updated.clear();
		rep2() {
			ivec2 pos = {i, j};
			int posi = pos.pack(N);
			for (int place : p->stationToPlaces[posi]) {
				dpincomes[posi] += p->pairDists[place >> 1];
			}
		}
	}

	void collapse(rngen& rng) {
		while (true) {
			bool done = true;
			for (auto& s : stations) {
				for (auto& c : s.connections) {
					if (!c.dirFixed()) {
						s.collapse(stations, updated, c, rng);
						done = false;
						update();
					}
				}
			}
			if (done)
				break;
		}
	}

	int finalScore() const {
		return money < 0 || turn > T ? -INF : money + (T - turn) * income;
	}

	void placeStation(const StationPlaceInfo& info) {
		auto [pos, score, railsToBePlaced, index1, index2] = info;
		assert(score >= 0);
		assert(index1 != -1 || stations.empty());
		auto& s = stations.emplace_back(pos, len(stations));
		if (index1 != -1) {
			if (index2 == -1) {
				// connect
				Station::connect(stations, updated, index1, s.index);
			} else {
				// insert
				Station::insert(stations, updated, index1, index2, s.index);
			}
		}

		update();

		auto waitForMoney = [&](int target) {
			if (money >= target)
				return;
			assert(income > 0);
			int toWait = (target - money + income - 1) / income;
			money += toWait * income;
			turn += toWait;
			assert(money >= target);
		};

		// place rails
		rep0(i, railsToBePlaced) {
			waitForMoney(RAIL_COST);
			money += income - RAIL_COST;
			turn++;
		}

		// place station
		waitForMoney(STATION_COST);
		money -= STATION_COST;

		// update dists and incomes and stuff
		afterStationPlacedAt(pos);

		// get new income
		money += income;
		turn++;
	}

	StationPlaceInfo computeNextPlaceInfo(ivec2 pos, bool upperBound) const {
		int posi = pos.pack(N);
		if (dists[posi] == 0)
			return {pos}; // occupied
		if (stations.empty()) {
			// the first one does not require rails
			return {pos, money < STATION_COST ? -INF : money - STATION_COST, 0};
		}
		int numRailsToBePlaced;
		short index1 = -1;
		short index2 = -1;
		if (upperBound) {
			numRailsToBePlaced = insertable[posi] ? 0 : dists[posi] - 1;
		} else {
			auto [numRails, si1, si2] = computeNumRailsAndIndices(pos);
			if (numRails == -1) {
				return {pos}; // fail
			}
			numRailsToBePlaced = numRails;
			index1 = si1;
			index2 = si2;
		}
		if (income == 0 && money < RAIL_COST * numRailsToBePlaced + STATION_COST) {
			return {pos}; // can't afford the next
		}

		int money = this->money;
		int income = this->income;
		int pincome = this->pincome;
		int turn = this->turn;

		// note: this is not exactly accurate
		int waitForRails =
			income == 0 ? 0 : max(0, (numRailsToBePlaced * RAIL_COST - money + income - 1) / income);
		int turnsForRails = max(numRailsToBePlaced, waitForRails + 1);
		// palce rails
		money += income * turnsForRails - numRailsToBePlaced * RAIL_COST;
		turn += turnsForRails;

		// this one is accurate though
		int waitForStation = income == 0 ? 0 : max(0, (STATION_COST - money + income - 1) / income);
		// place station
		int dincome = dincomes[posi];
		money += income * (waitForStation + 1) + dincome - STATION_COST;
		turn += waitForStation + 1;
		income += dincome;
		pincome += dpincomes[posi] - dincome;

		if (income == 0) {
			return {pos}; // no income after placing two or more stations
		}

		// fast forward to the next station placement
		int waitForNextStation = income == 0 ? money < STATION_COST ? -INF : 0
											 : max(0, (STATION_COST - money + income - 1) / income);
		waitForNextStation++; // + 1 for the placement itself
		if (turn + waitForNextStation > T) {
			return {pos}; // too late
		}
		turn += waitForNextStation;
		money += income * waitForNextStation - STATION_COST;

		if (turn > T) {
			return {pos}; // too late
		}
		money += (T - turn) * (income + pincome * partialIncomeCoeff());
		return {pos, money * powTable(max(T - 100 - turn, 0)), numRailsToBePlaced, index1, index2};
	}

	ull nextHash(ivec2 pos) const {
		ull res = hash;
		for (int place : p->stationToPlaces[pos.pack(N)]) {
			if (covered[place])
				continue;
			res ^= hashOfPlace(place);
		}
		return res;
	}

	void updateInsertableCells() {
		insertable.reset();
		for (auto& s1 : stations) {
			for (auto& c : s1.connections) {
				if (s1.index > c.index)
					continue; // avoid double counting
				auto& s2 = stations[c.index];
				auto [mins, maxs] = Station::connectionAabb(s1, s2);
				rep(i, mins.i, maxs.i + 1) {
					rep(j, mins.j, maxs.j + 1) {
						insertable[ivec2{i, j}.pack(N)] = true;
					}
				}
			}
		}
	}

	vector<pair<Tile, ivec2>> computeActions(bool checkOnly = false) {
		// determine final tiles
		array<Tile, N2> tiles;
		tiles.fill(Tile::empty());
		// reserved for rail ends
		bool reserved[N2] = {};
		for (auto& s : stations) {
			tiles[s.pos.pack(N)] = Tile::station();
			reserved[s.pos.pack(N)] = true;
		}
		static vector<pair<int, array<int, 2>>> stationPairs;
		stationPairs.clear();
		for (auto& s1 : stations) {
			for (auto& c12 : s1.connections) {
				if (s1.index > c12.index)
					continue; // avoid double counting
				auto& s2 = stations[c12.index];
				if ((s1.pos - s2.pos).mnorm() == 1)
					continue;
				auto& c21 = s2.getConnectionTo(s1.index);
				cvec2 p1 = s1.pos + cvec2::dir(c12.fixedDir());
				cvec2 p2 = s2.pos + cvec2::dir(c21.fixedDir());
				stationPairs.push_back({(p1 - p2).mnorm(), {s1.index, s2.index}});
				int p1i = p1.pack(N);
				int p2i = p2.pack(N);
				auto check = [&](int pi) {
					if (reserved[pi]) {
						trace("rail ends conflict!");
						return false;
					}
					reserved[pi] = true;
					return true;
				};
				if (!check(p1i) || (p1i != p2i && !check(p2i))) {
					return {};
				}
			}
		}
		// sort by distance
		ranges::sort(stationPairs);
		// process from the shortest
		for (auto [_, pair] : stationPairs) {
			auto [index1, index2] = pair;
			auto& s1 = stations[index1];
			auto& s2 = stations[index2];
			if ((s1.pos - s2.pos).mnorm() == 1)
				continue;
			auto& c12 = s1.getConnectionTo(index2);
			auto& c21 = s2.getConnectionTo(index1);
			auto [mins, maxs] = Station::connectionAabb(s1, s2);
			cvec2 p1 = s1.pos + cvec2::dir(c12.fixedDir());
			cvec2 p2 = s2.pos + cvec2::dir(c21.fixedDir());
			// bfs to fill the path
			array<int, N2> dist;
			array<cvec2, N2> prev;
			dist.fill(INF);
			prev.fill(cvec2{-1, -1});
			if (p1 != p2) {
				queue<cvec2> q;
				q.push(p1);
				dist[p1.pack(N)] = 0;
				[&]() {
					while (!q.empty()) {
						cvec2 pos = q.front();
						q.pop();
						int ndist = dist[pos.pack(N)] + 1;
						rep0(dir, 4) {
							cvec2 npos = pos + cvec2::dir(dir);
							int nposi = npos.pack_if_in_bounds(N);
							if (nposi == -1)
								continue;
							if (npos == p2) {
								prev[nposi] = pos;
								assert(dist[nposi] == INF);
								dist[nposi] = ndist;
								return;
							}
							if (dist[nposi] == INF && !reserved[nposi]) {
								dist[nposi] = ndist;
								prev[nposi] = pos;
								q.push(npos);
							}
						}
					}
				}();
				if (dist[p2.pack(N)] != (p1 - p2).mnorm()) {
					// failed to connect straightly
					trace("failed to connect straightly! ", s1.index, "-", s2.index);
					return {};
				}
			}
			// complement ends
			prev[p1.pack(N)] = s1.pos;
			prev[s2.pos.pack(N)] = p2;
			// restore path
			vector<cvec2> path;
			cvec2 pos = s2.pos;
			while (pos != s1.pos) {
				path.push_back(pos);
				pos = prev[pos.pack(N)];
			}
			path.push_back(s1.pos);
			reverse(path.begin(), path.end());
			// set tiles
			rep(i, 1, len(path) - 1) {
				cvec2 prev = path[i - 1];
				cvec2 cur = path[i];
				cvec2 next = path[i + 1];
				tiles[cur.pack(N)] = Tile::computeTile(prev, cur, next);
				reserved[cur.pack(N)] = true;
			}
		}
		if (checkOnly) {
			return {{Tile::empty(), ivec2{0, 0}}}; // return dummy
		}
		auto getStationPath = [&](int from, int to) {
			int numStations = stations.size();
			vector<int> prev(numStations, -1);
			vector<bool> visited(numStations, false);
			queue<int> q;
			q.push(from);
			visited[from] = true;
			while (!q.empty()) {
				int i = q.front();
				q.pop();
				if (i == to) {
					vector<int> path;
					while (i != -1) {
						path.push_back(i);
						i = prev[i];
					}
					reverse(path.begin(), path.end());
					return path;
				}
				for (auto& c : stations[i].connections) {
					int j = c.index;
					if (!visited[j]) {
						visited[j] = true;
						prev[j] = i;
						q.push(j);
					}
				}
			}
			mov.target("fail");
			draw();
			mov.comment("failed to find path from " + tos(from) + " to " + tos(to));
			mov.end_frame();
			assert(false);
		};

		// build actions
		vector<pair<Tile, ivec2>> actions;
		array<bool, N2> built;
		built.fill(false);
		auto build = [&](Tile t, ivec2 pos) {
			if (built[pos.pack(N)] && t != Tile::station()) {
				return;
			}
			actions.push_back({t, pos});
			built[pos.pack(N)] = true;
		};
		int numStations = stations.size();
		// build first station
		build(Tile::station(), stations[0].pos);
		// loop rails -> station -> ...
		rep(i, 1, numStations) {
			auto path = getStationPath(i - 1, i);
			rep(i, 1, len(path)) {
				int index1 = path[i - 1];
				int index2 = path[i];
				auto& s1 = stations[index1];
				auto& s2 = stations[index2];
				auto& c12 = s1.getConnectionTo(index2);
				auto& c21 = s2.getConnectionTo(index1);
				cvec2 p1 = s1.pos + cvec2::dir(c12.fixedDir());
				cvec2 p2 = s2.pos + cvec2::dir(c21.fixedDir());
				if ((s1.pos - s2.pos).mnorm() > 1) {
					cvec2 prev = s1.pos;
					cvec2 pos = p1;
					while (pos != s2.pos) {
						Tile t = tiles[pos.pack(N)];
						build(t, pos);
						// move to next
						cvec2 next = [&]() {
							rep0(dir, 4) {
								if (t.connects(dir)) {
									cvec2 npos = pos + cvec2::dir(dir);
									if (npos != prev) {
										return npos;
									}
								}
							}
							assert(false);
						}();
						prev = pos;
						pos = next;
					}
				}
				// bridge station if needed
				if (i < len(path) - 1) {
					// interpolate tile
					cvec2 prev = p2;
					cvec2 pos = s2.pos;
					auto& c23 = s2.getConnectionTo(path[i + 1]);
					cvec2 next = s2.pos + cvec2::dir(c23.fixedDir());
					Tile t = Tile::computeTile(prev, pos, next);
					build(t, pos);
				}
			}
			// build station
			build(Tile::station(), stations[i].pos);
		}
		return actions;
	}

	void draw() const {
		bool hasHome[N2] = {};
		bool hasWork[N2] = {};
		bool isCovered[N2] = {};
		int m = p->m;
		rep0(i, m) {
			ivec2 home = p->poss[i][0];
			ivec2 workplace = p->poss[i][1];
			hasHome[home.pack(N)] = true;
			hasWork[workplace.pack(N)] = true;
		}
		for (auto& s : stations) {
			for (int place : p->stationToPlaces[s.pos.pack(N)]) {
				ivec2 ppos = p->poss[place / 2][place % 2];
				isCovered[ppos.pack(N)] = true;
			}
		}
		mov.no_stroke();
		rep2() {
			ivec2 pos = {i, j};
			if (hasHome[pos.pack(N)]) {
				mov.fill(0, 1, 1, isCovered[pos.pack(N)] ? 1 : 0.4);
				mov.circle(j + 0.2, i + 0.2, 0.15);
			}
			if (hasWork[pos.pack(N)]) {
				mov.fill(1, 0, 1, isCovered[pos.pack(N)] ? 1 : 0.4);
				mov.circle(j + 0.8, i + 0.8, 0.15);
			}
		}

		mov.stroke_weight(0.1);
		mov.text_size(0.8);
		mov.text_align(mov.CENTER);
		// draw connections
		for (auto& s1 : stations) {
			ivec2 pos1 = s1.pos;
			for (auto& c : s1.connections) {
				auto& s2 = stations[c.index];
				ivec2 pos2 = s2.pos;
				mov.stroke(0, 0, 1);
				mov.line(pos1.j + 0.5, pos1.i + 0.5, pos2.j + 0.5, pos2.i + 0.5);
				if (s1.index < s2.index) {
					auto [mins, maxs] = Station::connectionAabb(s1, s2);
					mov.stroke(0, 0, 1);
					mov.fill(0, 0, 1, 0.2);
					mov.rect(mins.j, mins.i, maxs.j - mins.j + 1, maxs.i - mins.i + 1);
				}
			}
		}
		// draw labels
		for (auto& s1 : stations) {
			ivec2 pos1 = s1.pos;
			for (auto& c : s1.connections) {
				auto& s2 = stations[c.index];
				ivec2 pos2 = s2.pos;
				ivec2 diff = pos2 - pos1;
				double len = diff.enorm();
				double shift = min(1.0, len * 0.3);
				double textPosI = pos1.i + diff.i * shift / len;
				double textPosJ = pos1.j + diff.j * shift / len;
				string dirs = "";
				rep0(dir, 4) {
					if (c.hasDir(dir)) {
						dirs += "^v<>"[dir];
					}
				}
				mov.fill(0.8);
				mov.text(dirs, textPosJ + 0.5, textPosI + 0.5);
			}
		}
		// draw stations
		for (auto& s : stations) {
			ivec2 pos = s.pos;
			mov.fill(0.75, 0, 0);
			mov.no_stroke();
			string tooltip = "station " + tos(s.index) + " at " + tos(s.pos) + "\n";
			tooltip += "connected to: " + tos([&]() {
				vector<int> res;
				for (auto& c : s.connections) {
					res.push_back(c.index);
				}
				return res;
			}());
			mov.tooltip(tooltip);
			mov.circle(pos.j + 0.5, pos.i + 0.5, 0.4);
			mov.no_tooltip();
			mov.fill(1);
			mov.text(tos(s.index), pos.j + 0.5, pos.i);
		}
		mov.comment("nodes: " + tos(stations.size()));
		mov.comment("money: " + tos(money));
		mov.comment("income: " + tos(income));
		mov.comment("turn: " + tos(turn));
		mov.comment("final score: " + tos(finalScore()));
	}

private:
	void update() {
		// update connections until it converges
		while (!updated.empty()) {
			static vector<int> updateNext;
			updateNext.clear();
			updateNext.insert(updateNext.end(), updated.begin(), updated.end());
			updated.clear();
			for (int index : updateNext) {
				Station& s = stations[index];
				bool successfull = s.propagate(stations, updated);
				assert(successfull);
			}
		}
	}

	void afterStationPlacedAt(cvec2 pos) {
		int posi = pos.pack(N);
		// update nearest station distances
		static cvec2 q[N2];
		int qh = 0;
		int qt = 0;
		q[qt++] = pos;
		dists[posi] = 0;
		while (qh < qt) {
			cvec2 pos = q[qh++];
			int ndist = dists[pos.pack(N)] + 1;
			rep0(dir, 4) {
				cvec2 npos = pos + cvec2::dir(dir);
				int nposi = npos.pack_if_in_bounds(N);
				if (nposi == -1)
					continue;
				if (ndist < dists[nposi]) {
					dists[nposi] = ndist;
					q[qt++] = npos;
				}
			}
		}
		// update incomes
		income += dincomes[posi];
		pincome += dpincomes[posi] - dincomes[posi];
		// update delta incomes and covered places
		for (int place : p->stationToPlaces[posi]) {
			if (covered[place])
				continue;
			covered[place] = true;
			hash ^= hashOfPlace(place);
			int pair = place >> 1;
			int pairValue = p->pairDists[pair];
			cvec2 placePos = p->poss[pair][place & 1];
			cvec2 otherPlacePos = p->poss[pair][place & 1 ^ 1];
			int nposi;
			// is the other one covered?
			if (covered[place ^ 1]) {
				// then deduct income from this place
				for (auto dpos : C_COVER_RANGE) {
					nposi = (placePos + dpos).pack_if_in_bounds(N);
					if (nposi != -1) {
						dincomes[nposi] -= pairValue;
					}
				}
			} else {
				// if not, deduct pincome from this place and add income to the other place
				for (auto dpos : C_COVER_RANGE) {
					nposi = (placePos + dpos).pack_if_in_bounds(N);
					if (nposi != -1) {
						dpincomes[nposi] -= pairValue;
					}
					nposi = (otherPlacePos + dpos).pack_if_in_bounds(N);
					if (nposi != -1) {
						dincomes[nposi] += pairValue;
						dpincomes[nposi] -= pairValue;
					}
				}
			}
		}
		assert(dincomes[posi] == 0);
		assert(dpincomes[posi] == 0);
	}

	// {numRailsToBePlaced, index1, index2 (-1 for connection)}
	array<int, 3> computeNumRailsAndIndices(cvec2 pos) const {
		int posi = pos.pack(N);
		assert(dists[posi] > 0);
		assert(!stations.empty());

		if (insertable[posi]) { // apparently insertable, check if it is really insertable
			double minDist2ToLine = INF;
			array<int, 2> bestIndices = {-1, -1};
			for (auto& s1 : stations) {
				for (auto& c : s1.connections) {
					if (s1.index > c.index)
						continue; // avoid double counting
					auto& s2 = stations[c.index];
					if (Station::insertable(s1, s2, pos)) {
						auto p1 = s1.pos;
						auto p2 = s2.pos;
						// compute square distance to the line (not segment) p1-p2
						auto v1 = p2 - p1;
						auto v2 = pos - p1;
						double a = v1.i * v2.j - v1.j * v2.i;
						double len2 = v1.i * v1.i + v1.j * v1.j;
						double dist2 = a * a / len2;
						if (update_min(minDist2ToLine, dist2)) {
							bestIndices = {s1.index, s2.index};
						}
					}
				}
			}
			if (bestIndices[0] != -1) {
				auto [si1, si2] = bestIndices;
				bool ok = true;
				// this rarely fails tho...
				[&]() {
					ivec2 pos1 = stations[si1].pos;
					ivec2 pos2 = stations[si2].pos;
					for (auto& s1 : stations) {
						for (auto& c : s1.connections) {
							if (s1.index > c.index)
								continue; // avoid double counting
							auto& s2 = stations[c.index];
							if (segmentsIntersect(s1.pos, s2.pos, pos, pos1) ||
								segmentsIntersect(s1.pos, s2.pos, pos, pos2)) {
								ok = false;
								return;
							}
						}
					}
				}();
				if (ok) {
					return {0, si1, si2};
				}
			}
		}
		// connect to the nearest possible station
		int nearestIndex = -1;
		for (auto& s : stations) {
			if ((s.pos - pos).mnorm() == dists[posi]) {
				nearestIndex = s.index;
				break;
			}
		}
		assert(nearestIndex != -1);

		int minDist = INF;
		int minIndex = -1;
		auto tryStation = [&](int index) {
			auto& s = stations[index];
			if (s.connectable(pos)) {
				int dist = (s.pos - pos).mnorm();
				if (dist < minDist) {
					// check if segments don't intersect
					ivec2 a1 = s.pos;
					ivec2 b1 = pos;
					for (auto& s1 : stations) {
						ivec2 a2 = s1.pos;
						for (auto& c : s1.connections) {
							auto& s2 = stations[c.index];
							if (s1.index == s.index || s2.index == s.index)
								continue;
							if (s1.index < s2.index) {
								ivec2 b2 = s2.pos;
								if (segmentsIntersect(a1, b1, a2, b2)) {
									return; // fail!
								}
							}
						}
					}
					minDist = dist;
					minIndex = s.index;
				}
			}
		};

		// try the nearest one first; very likely to succeed
		tryStation(nearestIndex);
		if (minIndex != -1) {
			return {minDist - 1, minIndex, -1};
		}

		// try others
		for (auto& s : stations) {
			if (s.index != nearestIndex) {
				tryStation(s.index);
			}
		}
		if (minIndex != -1) {
			return {minDist - 1, minIndex, -1};
		}
		// failed to connect to any station
		return {-1, -1, -1};
	}
};

struct StationConnector {
	Stations stations;
	fast_iset<MAX_STATIONS> updated;

	array<bool, N2> insertable;

	// INF if forbidden
	array<int, N2> distToNearestStation;
	// O(|stations|) for each query, or O(N^2) for precomputation and O(1) for each query
	bool bruteForceSearch = false;

	StationConnector() {
	}

	void clear() {
		stations.clear();
		updated.clear();
		insertable.fill(false);
		distToNearestStation.fill(INF);
	}

	void updateInsertableCells() {
		insertable.fill(false);
		for (auto& s1 : stations) {
			for (auto& c : s1.connections) {
				if (s1.index > c.index)
					continue; // avoid double counting
				auto& s2 = stations[c.index];
				auto [mins, maxs] = Station::connectionAabb(s1, s2);
				rep(i, mins.i, maxs.i + 1) {
					rep(j, mins.j, maxs.j + 1) {
						insertable[ivec2{i, j}.pack(N)] = true;
					}
				}
			}
		}
	}

	void updateDistToNearestStation() {
		constexpr int BIG = N * 2;
		distToNearestStation.fill(BIG);

		if (!bruteForceSearch) {
			// bfs from stations
			static queue<ivec2> q;
			assert(q.empty());
			for (auto& s : stations) {
				q.push(s.pos);
				distToNearestStation[s.pos.pack(N)] = 0;
			}
			while (!q.empty()) {
				ivec2 pos = q.front();
				q.pop();
				int dist = distToNearestStation[pos.pack(N)];
				rep0(dir, 4) {
					ivec2 npos = pos + ivec2::dir(dir);
					int nposi = npos.pack_if_in_bounds(N);
					if (nposi == -1)
						continue;
					if (distToNearestStation[nposi] == BIG) {
						distToNearestStation[nposi] = dist + 1;
						q.push(npos);
					}
				}
			}
		}

		// forbid cells around stations
		for (auto& s : stations) {
			distToNearestStation[s.pos.pack(N)] = INF;
			// rep(di, -1, 2) {
			// 	rep(dj, -1, 2) {
			// 		ivec2 pos = s.pos + ivec2{di, dj};
			// 		int posi = pos.pack_if_in_bounds(N);
			// 		if (posi != -1) {
			// 			distToNearestStation[posi] = INF;
			// 		}
			// 	}
			// }
		}
	}

	// -1 if impossible
	int additionalRailsFor(cvec2 pos, bool upper = false) const {
		int posi = pos.pack(N);
		if (stations.empty()) {
			return 0;
		}
		if (distToNearestStation[posi] == INF) {
			return -1; // forbidden; too close to a station
		}
		// insert if possible
		if (insertable[posi]) {
			return 0;
		}
		if (upper) {
			if (bruteForceSearch) {
				int minDist = INF;
				for (auto& s : stations) {
					update_min(minDist, (s.pos - pos).mnorm());
				}
				assert(minDist != INF);
				return minDist - 1;
			} else {
				return distToNearestStation[posi] - 1;
			}
		}
		// connect to the nearest possible station
		int minDist = INF;
		int minIndex = -1;
		for (auto& s : stations) {
			int dist = (s.pos - pos).mnorm();
			if (dist < minDist) {
				if (s.connectable(pos)) {
					// check if segments intersect
					// TODO: optimize
					bool intersect = false;

					[&]() {
						ivec2 a1 = s.pos;
						ivec2 b1 = pos;
						for (auto& s1 : stations) {
							ivec2 a2 = s1.pos;
							for (auto& c : s1.connections) {
								auto& s2 = stations[c.index];
								if (s1.index == s.index || s2.index == s.index)
									continue;
								if (s1.index < s2.index) {
									ivec2 b2 = s2.pos;
									if (segmentsIntersect(a1, b1, a2, b2)) {
										intersect = true;
										return;
									}
								}
							}
						}
					}();

					if (!intersect) {
						minDist = dist;
						minIndex = s.index;
					}
				}
			}
		}
		if (minIndex == -1) {
			return -1;
		}
		return minDist - 1;
	}

	bool placeStations(const vector<cvec2>& poss, const vector<int>& parents, rngen& rng,
		const fast_iset<MAX_STATIONS>& mask, bool comment = false) {
		stations.clear();
		rep0(i, len(poss)) {
			stations.emplace_back(poss[i], i);
		}
		updated.clear();
		rep0(i, len(poss)) {
			if (mask.contains(i) && mask.contains(parents[i])) {
				if (parents[i] != -1) {
					if (!Station::connect(stations, updated, parents[i], i, true)) {
						if (comment)
							trace("failed to connect ", parents[i], " -> ", i, " (", poss[parents[i]], " -> ",
								poss[i], ")");
						return false;
					}
					if (!update(true)) {
						if (comment)
							trace("failed to update ", i);
						return false;
					}
				}
			}
		}
		return true;
	}

	bool placeStations(
		const vector<cvec2>& poss, const vector<int>& parents, rngen& rng, bool comment = false) {
		stations.clear();
		rep0(i, len(poss)) {
			stations.emplace_back(poss[i], i);
		}
		updated.clear();
		rep0(i, len(poss)) {
			if (parents[i] != -1) {
				if (!Station::connect(stations, updated, parents[i], i, true)) {
					if (comment)
						trace("failed to connect ", parents[i], " -> ", i, " (", poss[parents[i]], " -> ",
							poss[i], ")");
					return false;
				}
				if (!update(true)) {
					if (comment)
						trace("failed to update ", i);
					return false;
				}
			}
		}
		return true;
	}

	int placeStation(cvec2 pos) {
		int additionalRails = [&]() {
			if (stations.empty()) {
				Station& s = stations.emplace_back(pos, 0);
				return 0;
			}
			// insert if possible
			double minDist2ToLine = INF;
			array<int, 2> bestInsertIndex = {-1, -1};
			for (auto& s1 : stations) {
				for (auto& c : s1.connections) {
					if (s1.index > c.index)
						continue; // avoid double counting
					auto& s2 = stations[c.index];
					if (Station::insertable(s1, s2, pos)) {
						cvec2 p1 = s1.pos;
						cvec2 p2 = s2.pos;
						// compute square distance to the line (not segment) p1-p2
						double dist2 = [&]() {
							cvec2 v1 = p2 - p1;
							cvec2 v2 = pos - p1;
							double a = v1.i * v2.j - v1.j * v2.i;
							double len2 = v1.i * v1.i + v1.j * v1.j;
							return a * a / len2;
						}();
						if (update_min(minDist2ToLine, dist2)) {
							bestInsertIndex = {s1.index, s2.index};
						}
					}
				}
			}
			if (bestInsertIndex[0] != -1) {
				auto [si1, si2] = bestInsertIndex;
				bool ok = true;
				// this rarely fails tho...
				[&]() {
					ivec2 pos1 = stations[si1].pos;
					ivec2 pos2 = stations[si2].pos;
					for (auto& s1 : stations) {
						for (auto& c : s1.connections) {
							if (s1.index > c.index)
								continue; // avoid double counting
							auto& s2 = stations[c.index];
							if (segmentsIntersect(s1.pos, s2.pos, pos, pos1) ||
								segmentsIntersect(s1.pos, s2.pos, pos, pos2)) {
								ok = false;
								return;
							}
						}
					}
				}();
				if (ok) {
					Station& s3 = stations.emplace_back(pos, stations.size());
					Station::insert(stations, updated, si1, si2, s3.index);
					return 0;
				}
			}
			// connect to the nearest possible station
			int minDist = INF;
			int minIndex = -1;
			for (auto& s : stations) {
				if (s.connectable(pos)) {
					int dist = (s.pos - pos).mnorm();
					if (dist < minDist) {
						// check if segments intersect
						// TODO: optimize
						bool intersect = false;

						[&]() {
							ivec2 a1 = s.pos;
							ivec2 b1 = pos;
							for (auto& s1 : stations) {
								ivec2 a2 = s1.pos;
								for (auto& c : s1.connections) {
									auto& s2 = stations[c.index];
									if (s1.index == s.index || s2.index == s.index)
										continue;
									if (s1.index < s2.index) {
										ivec2 b2 = s2.pos;
										if (segmentsIntersect(a1, b1, a2, b2)) {
											intersect = true;
											return;
										}
									}
								}
							}
						}();
						if (!intersect) {
							minDist = dist;
							minIndex = s.index;
						}
					}
				}
			}
			if (minIndex == -1) {
				return -1;
			}
			Station& s = stations.emplace_back(pos, stations.size());
			Station::connect(stations, updated, minIndex, s.index);
			return minDist - 1;
		}();
		if (additionalRails == -1)
			return -1;
		update();
		return additionalRails;
	}

	void collapse(rngen& rng) {
		while (true) {
			bool done = true;
			for (auto& s : stations) {
				for (auto& c : s.connections) {
					if (!c.dirFixed()) {
						s.collapse(stations, updated, c, rng);
						done = false;
						update();
					}
				}
			}
			if (done)
				break;
		}
	}

	bool update(bool allowFail = false) {
		while (!updated.empty()) {
			static vector<int> toUpdate;
			toUpdate.clear();
			toUpdate.insert(toUpdate.end(), updated.begin(), updated.end());
			updated.clear();
			for (int index : toUpdate) {
				Station& s = stations[index];
				if (!s.propagate(stations, updated, allowFail)) {
					return false;
				}
			}
		}
		return true;
	}

	void draw() {
		mov.stroke_weight(0.1);
		mov.text_size(0.8);
		mov.text_align(mov.CENTER);
		// draw connections
		for (auto& s1 : stations) {
			ivec2 pos1 = s1.pos;
			for (auto& c : s1.connections) {
				auto& s2 = stations[c.index];
				ivec2 pos2 = s2.pos;
				mov.stroke(0, 0, 1);
				mov.line(pos1.j + 0.5, pos1.i + 0.5, pos2.j + 0.5, pos2.i + 0.5);
				if (s1.index < s2.index) {
					auto [mins, maxs] = Station::connectionAabb(s1, s2);
					mov.stroke(0, 0, 1);
					mov.fill(0, 0, 1, 0.2);
					mov.rect(mins.j, mins.i, maxs.j - mins.j + 1, maxs.i - mins.i + 1);
				}
			}
		}
		// draw labels
		for (auto& s1 : stations) {
			ivec2 pos1 = s1.pos;
			for (auto& c : s1.connections) {
				auto& s2 = stations[c.index];
				ivec2 pos2 = s2.pos;
				ivec2 diff = pos2 - pos1;
				double len = diff.enorm();
				double shift = min(1.0, len * 0.3);
				double textPosI = pos1.i + diff.i * shift / len;
				double textPosJ = pos1.j + diff.j * shift / len;
				string dirs = "";
				rep0(dir, 4) {
					if (c.hasDir(dir)) {
						dirs += "^v<>"[dir];
					}
				}
				mov.fill(1);
				mov.text(dirs, textPosJ + 0.5, textPosI + 0.5);
			}
		}
		// draw stations
		for (auto& s : stations) {
			ivec2 pos = s.pos;
			mov.fill(0.75, 0, 0);
			mov.no_stroke();
			mov.tooltip("station " + tos(s.index) + " at " + tos(s.pos));
			mov.circle(pos.j + 0.5, pos.i + 0.5, 0.4);
			mov.no_tooltip();
			mov.fill(1);
			mov.text(tos(s.index), pos.j + 0.5, pos.i);
		}
		mov.comment("nodes: " + tos(stations.size()));
	}

	vector<pair<Tile, ivec2>> computeActions(bool checkOnly = false) {
		// determine final tiles
		array<Tile, N2> tiles;
		tiles.fill(Tile::empty());
		// reserved for rail ends
		bool reserved[N2] = {};
		for (auto& s : stations) {
			tiles[s.pos.pack(N)] = Tile::station();
			reserved[s.pos.pack(N)] = true;
		}
		static vector<pair<int, array<int, 2>>> stationPairs;
		stationPairs.clear();
		for (auto& s1 : stations) {
			for (auto& c12 : s1.connections) {
				if (s1.index > c12.index)
					continue; // avoid double counting
				auto& s2 = stations[c12.index];
				if ((s1.pos - s2.pos).mnorm() == 1)
					continue;
				auto& c21 = s2.getConnectionTo(s1.index);
				cvec2 p1 = s1.pos + cvec2::dir(c12.fixedDir());
				cvec2 p2 = s2.pos + cvec2::dir(c21.fixedDir());
				stationPairs.push_back({(p1 - p2).mnorm(), {s1.index, s2.index}});
				int p1i = p1.pack(N);
				int p2i = p2.pack(N);
				auto check = [&](int pi) {
					if (reserved[pi]) {
						trace("rail ends conflict!");
						return false;
					}
					reserved[pi] = true;
					return true;
				};
				if (!check(p1i) || (p1i != p2i && !check(p2i))) {
					return {};
				}
			}
		}
		// sort by distance
		ranges::sort(stationPairs);
		// process from the shortest
		for (auto [_, pair] : stationPairs) {
			auto [index1, index2] = pair;
			auto& s1 = stations[index1];
			auto& s2 = stations[index2];
			if ((s1.pos - s2.pos).mnorm() == 1)
				continue;
			auto& c12 = s1.getConnectionTo(index2);
			auto& c21 = s2.getConnectionTo(index1);
			auto [mins, maxs] = Station::connectionAabb(s1, s2);
			cvec2 p1 = s1.pos + cvec2::dir(c12.fixedDir());
			cvec2 p2 = s2.pos + cvec2::dir(c21.fixedDir());
			// bfs to fill the path
			array<int, N2> dist;
			array<cvec2, N2> prev;
			dist.fill(INF);
			prev.fill(cvec2{-1, -1});
			if (p1 != p2) {
				queue<cvec2> q;
				q.push(p1);
				dist[p1.pack(N)] = 0;
				[&]() {
					while (!q.empty()) {
						cvec2 pos = q.front();
						q.pop();
						int ndist = dist[pos.pack(N)] + 1;
						rep0(dir, 4) {
							cvec2 npos = pos + cvec2::dir(dir);
							int nposi = npos.pack_if_in_bounds(N);
							if (nposi == -1)
								continue;
							if (npos == p2) {
								prev[nposi] = pos;
								assert(dist[nposi] == INF);
								dist[nposi] = ndist;
								return;
							}
							if (dist[nposi] == INF && !reserved[nposi]) {
								dist[nposi] = ndist;
								prev[nposi] = pos;
								q.push(npos);
							}
						}
					}
				}();
				if (dist[p2.pack(N)] != (p1 - p2).mnorm()) {
					// failed to connect straightly
					trace("failed to connect straightly! ", s1.index, "-", s2.index);
					return {};
				}
			}
			// complement ends
			prev[p1.pack(N)] = s1.pos;
			prev[s2.pos.pack(N)] = p2;
			// restore path
			vector<cvec2> path;
			cvec2 pos = s2.pos;
			while (pos != s1.pos) {
				path.push_back(pos);
				pos = prev[pos.pack(N)];
			}
			path.push_back(s1.pos);
			reverse(path.begin(), path.end());
			// set tiles
			rep(i, 1, len(path) - 1) {
				cvec2 prev = path[i - 1];
				cvec2 cur = path[i];
				cvec2 next = path[i + 1];
				tiles[cur.pack(N)] = Tile::computeTile(prev, cur, next);
				reserved[cur.pack(N)] = true;
			}
		}
		if (checkOnly) {
			return {{Tile::empty(), ivec2{0, 0}}}; // return dummy
		}
		auto getStationPath = [&](int from, int to) {
			int numStations = stations.size();
			vector<int> prev(numStations, -1);
			vector<bool> visited(numStations, false);
			queue<int> q;
			q.push(from);
			visited[from] = true;
			while (!q.empty()) {
				int i = q.front();
				q.pop();
				if (i == to) {
					vector<int> path;
					while (i != -1) {
						path.push_back(i);
						i = prev[i];
					}
					reverse(path.begin(), path.end());
					return path;
				}
				for (auto& c : stations[i].connections) {
					int j = c.index;
					if (!visited[j]) {
						visited[j] = true;
						prev[j] = i;
						q.push(j);
					}
				}
			}
			mov.target("fail");
			draw();
			mov.comment("failed to find path from " + tos(from) + " to " + tos(to));
			mov.end_frame();
			assert(false);
		};

		// build actions
		vector<pair<Tile, ivec2>> actions;
		array<bool, N2> built;
		built.fill(false);
		auto build = [&](Tile t, ivec2 pos) {
			if (built[pos.pack(N)] && t != Tile::station()) {
				return;
			}
			actions.push_back({t, pos});
			built[pos.pack(N)] = true;
		};
		int numStations = stations.size();
		// build first station
		build(Tile::station(), stations[0].pos);
		// loop rails -> station -> ...
		rep(i, 1, numStations) {
			auto path = getStationPath(i - 1, i);
			rep(i, 1, len(path)) {
				int index1 = path[i - 1];
				int index2 = path[i];
				auto& s1 = stations[index1];
				auto& s2 = stations[index2];
				auto& c12 = s1.getConnectionTo(index2);
				auto& c21 = s2.getConnectionTo(index1);
				cvec2 p1 = s1.pos + cvec2::dir(c12.fixedDir());
				cvec2 p2 = s2.pos + cvec2::dir(c21.fixedDir());
				if ((s1.pos - s2.pos).mnorm() > 1) {
					cvec2 prev = s1.pos;
					cvec2 pos = p1;
					while (pos != s2.pos) {
						Tile t = tiles[pos.pack(N)];
						build(t, pos);
						// move to next
						cvec2 next = [&]() {
							rep0(dir, 4) {
								if (t.connects(dir)) {
									cvec2 npos = pos + cvec2::dir(dir);
									if (npos != prev) {
										return npos;
									}
								}
							}
							assert(false);
						}();
						prev = pos;
						pos = next;
					}
				}
				// bridge station if needed
				if (i < len(path) - 1) {
					// interpolate tile
					cvec2 prev = p2;
					cvec2 pos = s2.pos;
					auto& c23 = s2.getConnectionTo(path[i + 1]);
					cvec2 next = s2.pos + cvec2::dir(c23.fixedDir());
					Tile t = Tile::computeTile(prev, pos, next);
					build(t, pos);
				}
			}
			// build station
			build(Tile::station(), stations[i].pos);
		}
		return actions;
	}
};

struct CompactState {
	Stations stations;
	short lastTurn = 0;
	int lastMoney = 0;
	int lastIncome = 0;
};

struct BsState_old { // mostly unnecessary, but no time for refactoring :(
	StationConnector sc;
	// whether the ith pair of home and workplace is connected
	vector<bool> pairFulfilled;
	// whether the place is covered by a station
	// note: place = i * 2 for ith home, i * 2 + 1 for ith workplace
	vector<bool> placeCovered;

	int numPairsFulfilled = 0;

	int m = 0;
	int initialMoney = 0;

	// the turn that the final station is built
	int lastTurn = 0;
	// the money at the last turn
	int lastMoney = 0;
	// the income at the last turn
	int lastIncome = 0;

	int lastPotentialIncome = 0;

	const Problem* p = nullptr;
	array<ull, MAX_M * 2> hashForPlace;
	ull hash = 0;

	double powTable[T + 1];

	BsState_old() {
	}

	void setup(const Problem& p, rngen& rng) {
		this->p = &p;
		m = p.m;
		initialMoney = p.k;
		repm(i) {
			ivec2 home = p.poss[i][0];
			ivec2 workplace = p.poss[i][1];
			hashForPlace[i * 2] = rng.next_ull();
			hashForPlace[i * 2 + 1] = rng.next_ull();
		}
		double powBase = 1.001 + m / 1600.0 * 0.008;
		rep0(t, T + 1) {
			powTable[t] = pow(powBase, t);
		}
		clear();
	}

	void clear() {
		sc.clear();
		pairFulfilled.assign(m, false);
		placeCovered.assign(m * 2, false);
		numPairsFulfilled = 0;
		lastTurn = 0;
		lastMoney = initialMoney;
		lastIncome = 0;
	}

	void exportTo(CompactState& cs) {
		cs.stations.copyFrom(sc.stations);
		cs.lastTurn = lastTurn;
		cs.lastMoney = lastMoney;
		cs.lastIncome = lastIncome;
	}

	void importFrom(const CompactState& cs) {
		sc.stations.copyFrom(cs.stations);
		lastTurn = cs.lastTurn;
		lastMoney = cs.lastMoney;
		lastIncome = cs.lastIncome;
		// recompute covered places and fulfilled pairs
		pairFulfilled.assign(m, false);
		placeCovered.assign(m * 2, false);
		numPairsFulfilled = 0;
		for (auto& s : sc.stations) {
			for (int place : p->stationToPlaces[s.pos.pack(N)]) {
				if (!placeCovered[place]) {
					placeCovered[place] = true;
					if (placeCovered[place ^ 1]) {
						int pair = place >> 1;
						pairFulfilled[pair] = true;
						numPairsFulfilled++;
					}
				}
			}
		}
	}

	void prepare(bool bruteForceSearch) {
		this->sc.bruteForceSearch = bruteForceSearch;
		sc.updateInsertableCells();
		sc.updateDistToNearestStation();
		hash = 0;
		lastPotentialIncome = 0;
		rep0(i, m * 2) {
			if (placeCovered[i]) {
				hash ^= hashForPlace[i];
				if ((i & 1) == 0 && !placeCovered[i + 1]) {
					lastPotentialIncome += p->pairDists[i >> 1];
				}
			}
		}
	}

	ull nextHashFor(ivec2 pos) {
		ull res = hash;
		for (int place : p->stationToPlaces[pos.pack(N)]) {
			if (!placeCovered[place]) {
				res ^= hashForPlace[place];
			}
		}
		return res;
	}

	ull initialHashFor(ivec2 pos1, ivec2 pos2) {
		ull res = 0;
		bool marked[MAX_M * 2] = {};
		for (ivec2 pos : {pos1, pos2}) {
			for (int place : p->stationToPlaces[pos.pack(N)]) {
				if (!marked[place]) {
					res ^= hashForPlace[place];
					marked[place] = true;
				}
			}
		}
		return res;
	}

	double nextScoreFor(ivec2 pos, bool upper = false) {
		if (sc.stations.empty()) {
			return lastMoney - STATION_COST;
		}

		int additionalRails = sc.additionalRailsFor(pos, upper);
		if (additionalRails == -1) {
			return -INF;
		}
		int turn = lastTurn;
		int money = lastMoney;
		int income = lastIncome;

		if (income == 0 && money < RAIL_COST * additionalRails + STATION_COST) {
			return -INF;
		}

		int waitForRails =
			income == 0 ? 0 : max(0, (additionalRails * RAIL_COST - money + income - 1) / income);
		int turnsForRails = max(additionalRails, waitForRails + 1);
		// palce rails
		turn += turnsForRails;
		money += income * turnsForRails - additionalRails * RAIL_COST;

		int waitForStation = income == 0 ? 0 : max(0, (STATION_COST - money + income - 1) / income);
		turn += waitForStation;
		money += income * waitForStation;
		// place station
		turn++;
		// update income
		int potentialIncome = lastPotentialIncome;
		for (int place : p->stationToPlaces[pos.pack(N)]) {
			if (!placeCovered[place]) {
				int dist = p->pairDists[place >> 1];
				if (placeCovered[place ^ 1]) {
					// can be fulfilled
					income += dist;
					potentialIncome -= dist;
				} else {
					potentialIncome += dist;
				}
			}
		}
		money += income - STATION_COST;

		// fast forward to the next station building
		waitForStation = max(0, (STATION_COST - money + income - 1) / (income + 1));
		if (turn + waitForStation > T) {
			return -INF;
		}
		turn += waitForStation;
		money += income * waitForStation - STATION_COST;

		if (turn > T) {
			return -INF;
		}
		double potentialIncomeCoeff = lerp(0.05, 0.1, linearstep(11000, 12000, initialMoney));
		money += (T - turn) * (income + potentialIncome * potentialIncomeCoeff);
		return money * powTable[T - turn];
	}

	void placeStation(ivec2 pos) {
		int additionalRails = sc.placeStation(pos);
		assert(additionalRails != -1);
		rep0(_, additionalRails) {
			waitForMoney(RAIL_COST);
			lastMoney -= RAIL_COST;
			fastForward(1);
		}
		waitForMoney(STATION_COST);
		lastMoney -= STATION_COST;
		// fulfill pairs
		for (int place : p->stationToPlaces[pos.pack(N)]) {
			if (!placeCovered[place]) {
				placeCovered[place] = true;
				if (placeCovered[place ^ 1]) {
					int pair = place >> 1;
					pairFulfilled[pair] = true;
					lastIncome += p->pairDists[pair];
					numPairsFulfilled++;
				}
			}
		}
		fastForward(1);
	}

	void draw() {
		bool hasHome[N2] = {};
		bool hasWork[N2] = {};
		bool isCovered[N2] = {};
		rep0(i, m) {
			ivec2 home = p->poss[i][0];
			ivec2 workplace = p->poss[i][1];
			hasHome[home.pack(N)] = true;
			hasWork[workplace.pack(N)] = true;
		}
		for (auto& s : sc.stations) {
			for (int place : p->stationToPlaces[s.pos.pack(N)]) {
				ivec2 ppos = p->poss[place / 2][place % 2];
				isCovered[ppos.pack(N)] = true;
			}
		}
		mov.no_stroke();
		rep2() {
			ivec2 pos = {i, j};
			if (hasHome[pos.pack(N)]) {
				mov.fill(0, 1, 1, isCovered[pos.pack(N)] ? 1 : 0.4);
				mov.circle(j + 0.2, i + 0.2, 0.15);
			}
			if (hasWork[pos.pack(N)]) {
				mov.fill(1, 0, 1, isCovered[pos.pack(N)] ? 1 : 0.4);
				mov.circle(j + 0.8, i + 0.8, 0.15);
			}
		}
		sc.draw();
	}

private:
	void waitForMoney(int target) {
		if (lastMoney >= target)
			return;
		assert(lastIncome > 0);
		int turns = (target - lastMoney + lastIncome - 1) / lastIncome;
		fastForward(turns);
		assert(lastMoney >= target);
	}

	void fastForward(int turns) {
		lastTurn += turns;
		lastMoney += turns * lastIncome;
	}
};

string DEBUG_STR = "";

struct SaState {
	struct Edge {
		int to;
		int dist;

		friend ostream& operator<<(ostream& out, const Edge& e) {
			return out << "{to=" << e.to << ", dist=" << e.dist << "}";
		}
	};
	struct Node {
		ivec2 pos;
		int index;
		int subIndex;
		int height;
		Edge parent;
		fast_vector<Edge, 4> children;
		int time = 0;

		int numConnections() const {
			return len(children) + (parent.to != -1);
		}

		void eraseChild(int child) {
			assert(hasChild(child));
			rep0(i, len(children)) {
				if (children[i].to == child) {
					children.erase(children.begin() + i);
					return;
				}
			}
		}

		bool hasChild(int child) {
			rep0(i, len(children)) {
				if (children[i].to == child) {
					return true;
				}
			}
			return false;
		}

		Edge& getChild(int child) {
			assert(hasChild(child));
			rep0(i, len(children)) {
				if (children[i].to == child) {
					return children[i];
				}
			}
			assert(false);
		}

		friend ostream& operator<<(ostream& out, const Node& n) {
			return out << "{pos=" << n.pos << ", index=" << n.index << ", subIndex=" << n.subIndex
					   << ", parent=" << n.parent.to << ", children=" << tos(n.children) << "}";
		}
	};
	struct Place {
		int index;
		fast_vector<int, 16> coveredBy;
		int minCoveredBy = INF;

		friend ostream& operator<<(ostream& out, const Place& e) {
			return out << "{i=" << e.index << ", by=" << tos(e.coveredBy) << "}";
		}

		void clear() {
			coveredBy.clear();
			minCoveredBy = INF;
		}

		void addCoveredBy(int index) {
			assert([&]() { // check if already covered
				for (int i : coveredBy) {
					if (i == index) {
						return false;
					}
				}
				return true;
			}());
			coveredBy.push_back(index);
			minCoveredBy = min(minCoveredBy, index);
		}

		void removeCoveredBy(int index) {
			int at = -1;
			minCoveredBy = INF;
			rep0(i, len(coveredBy)) {
				if (coveredBy[i] == index) {
					at = i;
				} else {
					minCoveredBy = min(minCoveredBy, coveredBy[i]);
				}
			}
			assert(at != -1);
			coveredBy.erase(coveredBy.begin() + at);
		}
	};
	struct ScoreInfo {
		int money = 0;
		int income = 0;
		int turn = 0;

		int finalScore() const {
			return money < 0 || turn > T ? -INF : money + income * (T - turn);
		}

		bool operator==(const ScoreInfo& s) const {
			return money == s.money && income == s.income && turn == s.turn;
		}

		bool operator!=(const ScoreInfo& s) const {
			return money != s.money || income != s.income || turn != s.turn;
		}

		friend ostream& operator<<(ostream& out, const ScoreInfo& s) {
			return out << "{money=" << s.money << ", income=" << s.income << ", turn=" << s.turn
					   << ", final=" << s.finalScore() << "}";
		}
	};

	fast_vector<Node, MAX_STATIONS> nodes = {};
	ScoreInfo* scores;
	array<Place, MAX_M * 2> places = {};
	array<int, N2> nodeIndex = {};
	int indexToIgnore = -1;

	SaState() {
		scores = scoreData.data() + 1;
	}

	void setup(const Problem& p) {
		this->p = &p;
		m = p.m;
		nodes.clear();
		nodeIndex.fill(-1);
		rep0(i, m) {
			ivec2 home = p.poss[i][0];
			ivec2 workplace = p.poss[i][1];
			places[i * 2] = {i * 2};
			places[i * 2 + 1] = {i * 2 + 1};
		}
		scores[-1] = {p.k, 0, 0};
	}

	void init(const vector<cvec2>& poss, const vector<int>& parents) {
		assert(len(poss) == len(parents));
		nodes.clear();
		int numNodes = len(poss);
		rep0(i, numNodes) {
			auto& node = nodes.emplace_back();
			node.pos = poss[i];
			node.index = i;
			node.subIndex = i;
			node.parent = {parents[i], 0};
			node.children.clear();
		}
		rep0(index, numNodes) {
			auto& node = nodes[index];
			int pindex = node.parent.to;
			if (pindex != -1) {
				auto& pnode = nodes[pindex];
				int dist = (node.pos - pnode.pos).mnorm();
				node.parent.dist = dist;
				pnode.children.push_back({index, dist});
			}
		}
		repm(pair) {
			int at = max(places[pair * 2].minCoveredBy, places[pair * 2 + 1].minCoveredBy);
			if (at != INF) {
				dincomes[at] += p->pairDists[pair];
			}
		}
		// compute initial subIndex by dfs
		auto dfs = [&](auto dfs, int index) -> int {
			auto& node = nodes[index];
			int subIndex = index;
			for (auto& c : node.children) {
				subIndex = min(subIndex, dfs(dfs, c.to));
			}
			node.subIndex = subIndex;
			return subIndex;
		};
		dfs(dfs, 0);
		recomputeDeltaIncome();
		updateInfo();
	}

	void recomputeDeltaIncome() {
		rep0(i, len(nodes)) {
			dincomes[i] = 0;
		}
		rep0(i, m * 2) {
			places[i].clear();
		}
		nodeIndex.fill(-1);
		rep0(index, len(nodes)) {
			cover(index);
		}
	}

	bool hasNoTJunction() const {
		rep0(index, len(nodes)) {
			auto& node = nodes[index];
			int smallerCount = 0;
			for (auto& c : node.children) {
				if (nodes[c.to].subIndex < node.index) {
					smallerCount++;
				}
			}
			if (smallerCount > 1) {
				return false;
			}
		}
		return true;
	}

	bool hasNoIntersectionForNode(int index, bool includeChildren = false) const {
		// TODO: optimizable? dunno
		auto& node = nodes[index];
		ivec2 a1 = node.pos;
		if (index == 0) {
			if (includeChildren) {
				rep(i, 1, len(nodes)) {
					ivec2 a2 = nodes[i].pos;
					ivec2 b2 = nodes[nodes[i].parent.to].pos;
					for (auto& c : node.children) {
						ivec2 b1_ = nodes[c.to].pos;
						if (segmentsIntersect(a1, b1_, a2, b2))
							return false;
					}
				}
			}
		} else {
			ivec2 b1 = nodes[node.parent.to].pos;
			rep(i, 1, len(nodes)) {
				if (i == index)
					continue;
				ivec2 a2 = nodes[i].pos;
				ivec2 b2 = nodes[nodes[i].parent.to].pos;
				if (segmentsIntersect(a1, b1, a2, b2))
					return false;
				if (includeChildren) {
					for (auto& c : node.children) {
						ivec2 b1_ = nodes[c.to].pos;
						if (segmentsIntersect(a1, b1_, a2, b2))
							return false;
					}
				}
			}
		}
		return true;
	}

	bool hasNoIntersectionForSegments(const vector<array<ivec2, 2>>& segments) const {
		rep(i, 1, len(nodes)) {
			ivec2 a2 = nodes[i].pos;
			ivec2 b2 = nodes[nodes[i].parent.to].pos;
			for (auto [a1, b1] : segments) {
				if (segmentsIntersect(a1, b1, a2, b2))
					return false;
			}
		}
		return true;
	}

	bool hasNoSelfIntersections() const {
		rep(i1, 1, len(nodes)) {
			auto& n1 = nodes[i1];
			ivec2 a1 = n1.pos;
			ivec2 b1 = nodes[n1.parent.to].pos;
			rep(i2, 1, len(nodes)) {
				if (i1 == i2)
					continue;
				auto& n2 = nodes[i2];
				ivec2 a2 = n2.pos;
				ivec2 b2 = nodes[n2.parent.to].pos;
				if (segmentsIntersect(a1, b1, a2, b2)) {
					trace("cross detected: ", i1, " ", i2);
					trace(a1, "-", b1, " vs ", a2, "-", b2);
					trace(DEBUG_STR);
					return false;
				}
			}
		}
		return true;
	}

	// must update info later if returns true
	bool normalize() {
		int numNodes = len(nodes);
		auto tryToImprove = [&]() {
			bool updated = false;
			rep0(i1, numNodes) {
				auto& n1 = nodes[i1];
				if (n1.parent.to == -1)
					continue;
				auto& p1 = nodes[n1.parent.to];
				if ((n1.pos - p1.pos).mnorm() == 1)
					continue;
				// compute AABB
				ivec2 mins = n1.pos.min(p1.pos);
				ivec2 maxs = n1.pos.max(p1.pos);

				auto tryN2 = [&](int i2) {
					auto& n2 = nodes[i2];
					if (n2.parent.to == i1 || n1.parent.to == i2)
						return false;
					ivec2 pos = n2.pos;
					if (pos.i < mins.i || pos.i > maxs.i || pos.j < mins.j || pos.j > maxs.j)
						return false;
					// n2 is in the n1-p1 AABB
					int subIndex1 = n1.subIndex;
					int subIndex2 = n2.subIndex;
					if (subIndex1 == subIndex2)
						return false;
					if (n2.numConnections() == 4)
						return false;
					if (n2.index > n2.subIndex && n1.subIndex < n2.index) {
						// causes T-junction
						return false;
					}

					if (subIndex1 < subIndex2) {
						assert(n2.parent.to != -1);

						// if (render) {
						// 	mov.target("normalization");
						// 	draw(false);
						// 	mov.comment("n1=" + tos(n1.index) + " n2=" + tos(n2.index));
						// 	mov.end_frame();
						// }

						// insert n2 between n1 and p1
						auto& p2 = nodes[n2.parent.to];
						// erase connections
						p2.eraseChild(i2);
						p1.eraseChild(i1);
						int distn1n2 = (n2.pos - n1.pos).mnorm();
						int distn2p1 = (p1.pos - n2.pos).mnorm();
						// connect n2 to p1
						n2.parent = {p1.index, distn2p1};
						p1.children.push_back({i2, distn2p1});
						// connect n1 to n2
						n1.parent = {n2.index, distn1n2};
						n2.children.push_back({i1, distn1n2});
						// update
						int minSubIndex = min(subIndex1, subIndex2);
						updateFrom(minSubIndex);

						if (!hasNoIntersectionForNode(i1) || !hasNoIntersectionForNode(i2)) {
							// revert
							p1.eraseChild(i2);
							n2.eraseChild(i1);
							int distn1p1 = (p1.pos - n1.pos).mnorm();
							n1.parent = {p1.index, distn1p1};
							p1.children.push_back({i1, distn1p1});
							int distn2p2 = (p2.pos - n2.pos).mnorm();
							n2.parent = {p2.index, distn2p2};
							p2.children.push_back({i2, distn2p2});
							updateFrom(minSubIndex);
							return false;
						}

						// trace("improved by inserting ", i2, " between ", i1, " and ", n2.parent.to);
					} else { // subIndex1 > subIndex2

						// if (render) {
						// 	mov.target("normalization");
						// 	draw(false);
						// 	mov.comment("n1=" + tos(n1.index) + " n2=" + tos(n2.index));
						// 	mov.end_frame();
						// }

						// connect n1 to n2
						// erase connections
						p1.eraseChild(i1);
						int distn1n2 = (n2.pos - n1.pos).mnorm();
						n1.parent = {i2, distn1n2};
						n2.children.push_back({i1, distn1n2});
						// update
						updateFrom(subIndex1);
						// trace("improved by connecting ", i1, " to ", i2);

						if (!hasNoIntersectionForNode(i1)) {
							// revert
							n2.eraseChild(i1);
							int distn1p1 = (p1.pos - n1.pos).mnorm();
							n1.parent = {p1.index, distn1p1};
							p1.children.push_back({i1, distn1p1});
							updateFrom(subIndex1);
							return false;
						}
					}

					// if (render) {
					// 	// debug rendering
					// 	mov.target("normalization");
					// 	draw(false);
					// 	mov.comment("done!");
					// 	mov.end_frame();
					// }
					updated = true;
					return true;
				};

				int area = (maxs.i - mins.i + 1) * (maxs.j - mins.j + 1);
				if (area < numNodes) {
					[&]() {
						rep(i, mins.i, maxs.i + 1) {
							rep(j, mins.j, maxs.j + 1) {
								int i2 = nodeIndex[ivec2{i, j}.pack(N)];
								if (i2 != -1 && tryN2(i2))
									return;
							}
						}
					}();
				} else {
					rep0(i2, numNodes) {
						if (tryN2(i2))
							break;
					}
				}
			}
			return updated;
		};
		if (!tryToImprove()) {
			return false;
		}
		while (true) {
			if (!tryToImprove())
				break;
		}
		return true;
	}

	// must be called after topology changes
	void updateInfo() {
		static queue<int> q;
		assert(q.empty());
		q.push(0);
		nodes[0].height = 0;
		while (!q.empty()) {
			int index = q.front();
			q.pop();
			auto& node = nodes[index];
			for (auto& c : node.children) {
				auto& child = nodes[c.to];
				child.height = node.height + 1;
				q.push(c.to);
			}
		}
		nodePoss.clear();
		rep0(i, len(nodes)) {
			nodePoss.insert(nodes[i].pos.pack(N));
		}
	}

	int finalScore() const {
		return scores[len(nodes) - 1].finalScore();
	}

	ScoreInfo computeScoreForNodeInsertion(int nodeIndex, int cindex, ivec2 pos, bool debug = false) {
		assert(indexToIgnore == -1);
		assert(nodeIndex > 0); // cannot insert at 0
		assert(cindex > 0); // must have parent
		int pindex = nodes[cindex].parent.to;
		Node& child = nodes[cindex];
		Node& parent = nodes[pindex];
		// if true, no need to build a new rail when the new node is built
		bool putOnRail = child.subIndex < nodeIndex;
		// only trust `subIndex` that are smaller than `minIndex`
		int minIndex = min(nodeIndex, child.subIndex);

		// update the rail cost of the child
		int oldDist = child.parent.dist;
		int newDist;
		if (putOnRail) {
			// cindex [-> nodeIndex ->] pindex
			//               ^ cover the whole distance
			newDist = (child.pos - pos).mnorm() + (parent.pos - pos).mnorm();
		} else {
			// cindex -> nodeIndex [->] pindex
			//                      ^ only cover this distance
			newDist = (child.pos - pos).mnorm();
		}
		child.parent.dist = newDist;

		int dincomeForNewNode = 0;
		for (int place : p->stationToPlaces[pos.pack(N)]) {
			int at = places[place].minCoveredBy;
			int otherAt = places[place ^ 1].minCoveredBy;
			if (at >= nodeIndex && otherAt < nodeIndex) {
				dincomeForNewNode += p->pairDists[place >> 1];
			}
		}
		bool dirty = false;

		auto res = [&]() -> ScoreInfo {
			int index = minIndex;
			ScoreInfo score = scores[index - 1];
			clearMarks();
			while (true) { // nodeIndex may be equal to len(nodes)
				// insert node
				if (index == nodeIndex) {
					// num rails required to place this station
					int rails = 0;
					if (!putOnRail) {
						// cindex -> nodeIndex [->] pindex
						//                      ^ build this rail
						rails += (parent.pos - pos).mnorm();

						// cindex -> nodeIndex -> pindex [->] ...
						//                                ^ then build this rail if needed
						rails += getRailPathDistanceWhileMarking(pindex, minIndex, true);
					}
					// exclude the end
					rails = max(rails - 1, 0);

					// update score
					score = nextScore(score, rails, dincomeForNewNode);
					if (score.money < 0) {
						return {-1, 0, 0};
					}

					cover(-1, pos);
					dirty = true;
				}

				// break if end
				if (index >= len(nodes)) {
					assert(index == len(nodes));
					break;
				}

				// num rails required to place this station
				int rails = getRailPathDistanceWhileMarking(index, minIndex, true);
				// exclude the end
				rails = max(rails - 1, 0);

				// get delta income
				int dincome = dincomes[index];
				if (debug)
					trace("rails=", rails, " score=", score);

				// update score
				score = nextScore(score, rails, dincome);
				if (score.money < 0) {
					return {-1, 0, 0};
				}
				index++;
			}
			return score;
		}();

		// restore the rail cost of the child
		child.parent.dist = oldDist;

		// restore dincome
		if (dirty)
			uncover(-1, pos);

		return res;
	}

	ScoreInfo computeScoreForNodeConnection(int nodeIndex, int pindex, ivec2 pos, bool debug = false) {
		assert(indexToIgnore == -1);
		assert(nodeIndex > 0); // cannot insert at 0
		Node& parent = nodes[pindex];
		assert(parent.numConnections() < 4);
		assert(parent.subIndex == parent.index || nodeIndex > parent.subIndex); // forbid T-junction
		// only trust `subIndex` that are smaller than `minIndex`
		int minIndex = nodeIndex;
		int index = minIndex;

		int dincomeForNewNode = 0;
		for (int place : p->stationToPlaces[pos.pack(N)]) {
			int at = places[place].minCoveredBy;
			int otherAt = places[place ^ 1].minCoveredBy;
			if (at >= nodeIndex && otherAt < nodeIndex) {
				dincomeForNewNode += p->pairDists[place >> 1];
			}
		}
		bool dirty = false;

		auto res = [&]() -> ScoreInfo {
			ScoreInfo score = scores[index - 1];
			clearMarks();
			while (true) { // nodeIndex may be equal to len(nodes)
				// insert node
				if (index == nodeIndex) {
					// num rails required to place this station
					int rails = (parent.pos - pos).mnorm();
					rails += getRailPathDistanceWhileMarking(pindex, minIndex, true);
					// exclude the end
					rails = max(rails - 1, 0);

					// update score
					score = nextScore(score, rails, dincomeForNewNode);
					if (score.money < 0) {
						return {-1, 0, 0};
					}

					cover(-1, pos);
					dirty = true;
				}

				// break if end
				if (index >= len(nodes)) {
					assert(index == len(nodes));
					break;
				}

				// num rails required to place this station
				int rails = getRailPathDistanceWhileMarking(index, minIndex, true);
				// exclude the end
				rails = max(rails - 1, 0);

				// compute delta income
				int dincome = dincomes[index];
				if (debug)
					trace("rails=", rails, " score=", score);

				// update score
				score = nextScore(score, rails, dincome);
				if (score.money < 0) {
					return {-1, 0, 0};
				}
				index++;
			}
			return score;
		}();

		// restore dincome
		if (dirty)
			uncover(-1, pos);

		return res;
	}

	ScoreInfo updateFrom(int index, bool dry = false, bool debug = false) {
		// only trust `subIndex` that are smaller than `minIndex`
		int minIndex = index;
		ScoreInfo score = scores[index - 1];
		clearMarks();
		while (index < len(nodes)) {
			if (index == indexToIgnore) {
				index++;
				continue;
			}
			// num rails required to place this station
			int rails = getRailPathDistanceWhileMarking(index, minIndex, dry);
			// exclude the end
			rails = max(rails - 1, 0);

			// get delta income
			int dincome = dincomes[index];
			if (debug)
				trace("rails=", rails, " score=", score);

			// update score
			score = nextScore(score, rails, dincome);
			if (!dry) {
				assert(score.money >= 0);
				scores[index] = score;
			}
			if (score.money < 0) {
				return {-1, 0, 0};
			}
			index++;
		}
		return score;
	}

	void draw() {
		mov.text_align(mov.CENTER);
		mov.no_stroke();
		mov.text_size(0.8);
		rep0(i, len(nodes)) {
			auto& node = nodes[i];
			ivec2 pos = node.pos;
			for (auto& c : node.children) {
				ivec2 cpos = nodes[c.to].pos;
				mov.stroke(0, 0, 1);
				mov.line(pos.j + 0.5, pos.i + 0.5, cpos.j + 0.5, cpos.i + 0.5);
			}
		}
		rep0(i, len(nodes)) {
			auto& node = nodes[i];
			ivec2 pos = node.pos;
			string text = tos(i) + " (" + tos(node.subIndex) + ")";
			string detailedText = text + "\np=" + tos(node.parent) + "\nc=" + tos(node.children);
			mov.tooltip(detailedText);
			mov.fill(0, 0, 1);
			mov.circle(pos.j + 0.5, pos.i + 0.5, 0.4);
			mov.no_tooltip();
			mov.fill(1);
			mov.text(text, pos.j + 0.5, pos.i + 0.5);
		}
		mov.comment("nodes: " + tos(len(nodes)));
	}

	bool isOccupied(ivec2 pos) const {
		return nodePoss.contains(pos.pack(N));
	}

	void uncover(int index) {
		uncover(index, nodes[index].pos);
	}

	void cover(int index) {
		cover(index, nodes[index].pos);
	}

private:
	int m = 0;

	const Problem* p = nullptr;
	array<ScoreInfo, MAX_STATIONS + 1> scoreData = {};
	fast_iset<N2> nodePoss;

	array<fast_vector<int, 16>, N2> placeCoveredBy = {};
	array<int, MAX_STATIONS + 1> dincomes = {};

	int time = 0;

	void uncover(int index, ivec2 pos) {
		int posi = pos.pack(N);
		assert(nodeIndex[posi] == index);
		nodeIndex[posi] = -1;
		for (int place : p->stationToPlaces[posi]) {
			int prevAt = max(places[place].minCoveredBy, places[place ^ 1].minCoveredBy);
			if (prevAt != INF)
				dincomes[prevAt] -= p->pairDists[place >> 1];
			places[place].removeCoveredBy(index);
			int newAt = max(places[place].minCoveredBy, places[place ^ 1].minCoveredBy);
			if (newAt != INF)
				dincomes[newAt] += p->pairDists[place >> 1];
		}
	}

	void cover(int index, ivec2 pos) {
		int posi = pos.pack(N);
		assert(nodeIndex[posi] == -1);
		nodeIndex[posi] = index;
		for (int place : p->stationToPlaces[posi]) {
			int prevAt = max(places[place].minCoveredBy, places[place ^ 1].minCoveredBy);
			if (prevAt != INF)
				dincomes[prevAt] -= p->pairDists[place >> 1];
			places[place].addCoveredBy(index);
			int newAt = max(places[place].minCoveredBy, places[place ^ 1].minCoveredBy);
			if (newAt != INF)
				dincomes[newAt] += p->pairDists[place >> 1];
		}
	}

	void updateSubIndexFrom(int index) {
		int pindex = nodes[index].parent.to;
		while (pindex != -1 && nodes[pindex].subIndex > nodes[index].subIndex) {
			nodes[pindex].subIndex = nodes[index].subIndex;
			index = pindex;
			pindex = nodes[index].parent.to;
		}
	}

	bool nodeMarked(int index) {
		return nodes[index].time == time;
	}

	void markNode(int index) {
		nodes[index].time = time;
	}

	void clearMarks() {
		time++;
	}

	int getRailPathDistanceWhileMarking(int index, int minIndex, bool dry) {
		int res = 0;
		int baseIndex = index;
		while (baseIndex > 0 && nodes[baseIndex].subIndex >= minIndex && !nodeMarked(baseIndex)) {
			markNode(baseIndex);
			if (!dry)
				nodes[baseIndex].subIndex = index;
			auto [pindex, dist] = nodes[baseIndex].parent;
			// trace("building rails from ", baseIndex, " to ", pindex, " dist: ", dist);
			res += dist;
			baseIndex = pindex;
		}
		return res;
	}

	ScoreInfo nextScore(ScoreInfo base, int rails, int dincome) {
		auto [money, income, turn] = base;
		// build rails
		if (rails > 0) {
			if (money + (rails - 1) * income >= rails * RAIL_COST) {
				// can build without waiting
				money += (income - RAIL_COST) * rails;
				turn += rails;
				// trace("built ", rails, " rails. money: ", money);
			} else {
				// build rails one by one
				int done = 0;
				while (done < rails) {
					if (money < RAIL_COST) {
						if (income == 0)
							return {-1, 0, 0};
						int wait = (RAIL_COST - money + income - 1) / income;
						turn += wait;
						money += wait * income;
					}
					assert(money >= RAIL_COST);
					money += income - RAIL_COST;
					turn++;
					done++;
					// trace("built a rail. money: ", money);
				}
			}
		}
		// wait for station
		if (money < STATION_COST) {
			if (income == 0)
				return {-1, 0, 0};
			int wait = income == 0 ? 0 : (STATION_COST - money + income - 1) / income;
			money += wait * income;
			turn += wait;
		}
		// build station
		assert(money >= STATION_COST);
		income += dincome;
		money += income - STATION_COST;
		turn++;
		// trace("built a station. money: ", money, " income: ", income);
		return {money, income, turn};
	}
};

struct Verifier {
	static int computeScore(const Problem& p, const string& output) {
		bool pairFulfilled[MAX_M] = {};
		int numPairsFulfilled = 0;
		int money = p.k;
		int income = 0;
		int turn = 0;
		array<Tile, N2> tiles;
		tiles.fill(Tile::empty());
		istringstream in(output);
		string line;

		array<int, N2> groundGroups;
		int groundGroupCount = 0;

		auto updateGroundGroups = [&]() {
			groundGroups.fill(0);
			groundGroupCount = 0;
			rep2() {
				ivec2 pos = {i, j};
				int posi = pos.pack(N);
				if (groundGroups[posi] > 0)
					continue;
				if (tiles[posi] == Tile::empty())
					continue;
				groundGroupCount++;
				// bfs
				queue<ivec2> q;
				q.push(pos);
				groundGroups[posi] = groundGroupCount;
				while (!q.empty()) {
					ivec2 p = q.front();
					q.pop();
					int posi = p.pack(N);
					rep0(dir, 4) {
						ivec2 np = p + ivec2::dir(dir);
						int npi = np.pack_if_in_bounds(N);
						if (npi == -1)
							continue;
						if (tiles[posi].connects(dir) && tiles[npi].connects(dir ^ 1) &&
							groundGroups[npi] == 0) {
							groundGroups[npi] = groundGroupCount;
							q.push(np);
						}
					}
				}
			}
		};

		auto draw = [&]() {
			if (!render)
				return;
			mov.target("verify");
			updateGroundGroups();
			mov.stroke_weight(0.2);
			bool isHome[N2] = {};
			bool isWork[N2] = {};
			rep0(i, p.m) {
				ivec2 home = p.poss[i][0];
				ivec2 workplace = p.poss[i][1];
				isHome[home.pack(N)] = true;
				isWork[workplace.pack(N)] = true;
			}
			mov.no_stroke();
			rep2() {
				ivec2 pos = {i, j};
				if (isHome[pos.pack(N)]) {
					mov.fill(1, 0, 1);
					mov.circle(j + 0.2, i + 0.2, 0.15);
				}
				if (isWork[pos.pack(N)]) {
					mov.fill(0, 1, 1);
					mov.circle(j + 0.8, i + 0.8, 0.15);
				}
			}
			rep2() {
				ivec2 pos = {i, j};
				if (tiles[pos.pack(N)] == Tile::empty())
					continue;
				int group = groundGroups[pos.pack(N)];
				double angle = group * 2.4; // golden angle
				array<double, 3> rgb = {
					0.5 + 0.5 * cos(angle),
					0.5 + 0.5 * cos(angle + numbers::pi * 2 / 3),
					0.5 + 0.5 * cos(angle + numbers::pi * 4 / 3),
				};
				mov.stroke(rgb[0], rgb[1], rgb[2]);
				rep0(dir, 4) {
					if (tiles[pos.pack(N)].connects(dir)) {
						ivec2 pos2 = pos + ivec2::dir(dir);
						mov.line(pos.j + 0.5, pos.i + 0.5, (pos.j + pos2.j) * 0.5 + 0.5,
							(pos.i + pos2.i) * 0.5 + 0.5);
					}
				}
			}
			mov.comment("turn: " + tos(turn) + " money: " + tos(money) + " income: " + tos(income));
			mov.end_frame();
		};

		auto checkPairs = [&]() {
			updateGroundGroups();

			rep0(i, p.m) {
				if (pairFulfilled[i])
					continue;
				ivec2 home = p.poss[i][0];
				ivec2 workplace = p.poss[i][1];
				vector<int> homeStations;
				vector<int> workStations;
				rep(di, -2, 3) rep(dj, -2, 3) {
					if (abs(di) + abs(dj) > 2)
						continue;
					ivec2 pos = home + ivec2(di, dj);
					int posi = pos.pack_if_in_bounds(N);
					if (posi != -1 && tiles[posi] == Tile::station()) {
						homeStations.push_back(posi);
					}
					pos = workplace + ivec2(di, dj);
					posi = pos.pack_if_in_bounds(N);
					if (posi != -1 && tiles[posi] == Tile::station()) {
						workStations.push_back(posi);
					}
				}
				for (int h : homeStations) {
					for (int w : workStations) {
						if (groundGroups[h] == groundGroups[w]) {
							// trace("pair ", i, " fulfilled");
							pairFulfilled[i] = true;
							income += (home - workplace).mnorm();
							numPairsFulfilled++;
							break;
						}
					}
					if (pairFulfilled[i])
						break;
				}
			}
		};

		draw();
		while (getline(in, line)) {
			istringstream ss(line);
			int kind;
			ss >> kind;
			if (kind == -1) { // wait
				turn++;
				money += income;
			} else if (kind == 0) { // place station
				int i, j;
				ss >> i >> j;
				ivec2 pos = {i, j};
				if (money < STATION_COST) {
					trace("not enough money for station at ", pos, " money: ", money, " income: ", income,
						" turn: ", turn);
					return 0;
				}
				money -= STATION_COST;
				tiles[pos.pack(N)] = Tile::station();
				checkPairs();
				money += income;
				turn++;
			} else {
				int i, j;
				ss >> i >> j;
				ivec2 pos = {i, j};
				if (money < RAIL_COST) {
					trace("not enough money for rail at ", i, j, " money: ", money, " income: ", income,
						" turn: ", turn);
					return 0;
				}
				money -= RAIL_COST;
				tiles[pos.pack(N)] = [&]() {
					switch (kind) {
					case 1:
						return Tile::lr();
					case 2:
						return Tile::ud();
					case 3:
						return Tile::ld();
					case 4:
						return Tile::lu();
					case 5:
						return Tile::ur();
					case 6:
						return Tile::rd();
					default:
						return Tile::empty();
					}
				}();
				if (tiles[pos.pack(N)] == Tile::empty()) {
					trace("invalid rail type at ", pos, " kind: ", kind);
					return 0;
				}
				checkPairs();
				money += income;
				turn++;
			}
			draw();
		}
		if (turn != T) {
			trace("turns must be ", T, " but got ", turn);
			return 0;
		}
		return money;
	}
};

// weighted random integer selector
struct Selector {
	vector<double> weights;
	double totalWeight = 0;
	int last = 0;

	Selector() {
	}

	void clear() {
		weights.clear();
		totalWeight = 0;
	}

	void add(double weight) {
		weights.push_back(weight);
		totalWeight += weight;
	}

	int select(rngen& rng) {
		double r = rng.next_float(0, totalWeight);
		double sum = 0;
		rep0(i, len(weights)) {
			sum += weights[i];
			if (r < sum) {
				return last = i;
			}
		}
		// may happen due to floating point error
		return last = len(weights) - 1;
	}
};

struct Params {
	array<double, 2> trans0 = {1, 1};
	array<double, 2> trans1 = {1, 1};
	array<double, 2> trans2 = {1, 1};
	array<double, 2> trans3 = {1, 1};
	array<double, 2> trans4 = {1, 1};
	array<double, 2> trans5 = {1, 1};
	array<double, 2> trans6 = {1, 1};

	Params() {
		trans0[0] = 0.952769394492807;
		trans0[1] = 0.7888673621420587;
		trans1[0] = 0.31910378891292307;
		trans1[1] = 0.506937277844673;
		trans2[0] = 0.817336202936464;
		trans2[1] = 0.867783320771288;
		trans3[0] = 0.5818921125105023;
		trans3[1] = 0.6921190250102456;
		trans4[0] = 0.21302815714945808;
		trans4[1] = 0.8329906322773127;
		trans5[0] = 0.047384743222474796;
		trans5[1] = 0.34000115463345465;
		trans6[0] = 0.03618312075516136;
		trans6[1] = 0.10624169007119856;
	}
};

struct SolverResult {
	ll score = 0;
};

class Solver {
public:
	int seed;
	int m;
	int k;
	rngen rng;
	SolverResult result;
	Problem p;

	Params params;

	Solver() : rng(12345) {
	}

	void load(istream& in, int seed = -1) {
		p.load(in);
		m = p.m;
		k = p.k;
		mov.set_file(movie_file_name(seed));
		init();
	}

	void load(int seed) {
		this->seed = seed;
		istringstream in(read_text_assert(input_file_name(seed)));
		isLocal = true;
		load(in, seed);
	}

	void solve() {
		if (isLocal) {
			ostringstream out;
			solveMain(out);

			// int trueScore = Verifier::computeScore(p, out.str());
			// if (result.score != trueScore) {
			// 	trace("!!!!!!!!!!!!!!!!!!!!!!!!!");
			// 	trace("computed score: ", result.score, ", true score: ", trueScore);
			// }
			// assert(result.score == trueScore);
			// result.score = trueScore;

			mov.close_file();
			write_text(output_file_name(seed), out.str());
		} else {
			solveMain(cout);
		}
	}

private:
	void init() {
	}

	void solveMain(ostream& out) {
		// run beam search for the initial solution
		auto [poss, parents] = bs();

		// run SA
		sa(poss, parents, out);
	}

	static constexpr int MAX_BEAM_WIDTH = 10000;

	tuple<vector<cvec2>, vector<int>> bs() {
		BsState initialState;
		initialState.setup(p, rng);

		struct Candidate {
			int parent;
			StationPlaceInfo placeInfo;

			Candidate(int parent, StationPlaceInfo placeInfo) : parent(parent), placeInfo(placeInfo) {
			}

			Candidate() : parent(-1), placeInfo({-1, -1}) {
			}
		};

		// determine the first pair to connect
		// TODO: optimize?
		vector<array<ivec2, 2>> initialCands;
		bitset<N2 * N2> added;
		auto idOf = [&](ivec2 p1, ivec2 p2) {
			int pi1 = p1.pack(N);
			int pi2 = p2.pack(N);
			if (pi1 > pi2)
				swap(pi1, pi2);
			return pi1 * N2 + pi2;
		};
		int maxDist = (p.k - STATION_COST * 2) / RAIL_COST + 1;
		rep2() {
			ivec2 p1 = {i, j};
			for (int place : p.stationToPlaces[p1.pack(N)]) {
				if ((place & 1) == 1)
					continue;
				int pair = place >> 1;
				int other = place ^ 1;
				ivec2 otherPos = p.poss[pair][other & 1];
				rep(di, -2, 3) rep(dj, -2, 3) {
					if (abs(di) + abs(dj) > 2)
						continue;
					ivec2 p2 = otherPos + ivec2(di, dj);
					int dist = (p2 - p1).mnorm();
					if (dist > maxDist || dist <= 2)
						continue;
					int p2i = p2.pack_if_in_bounds(N);
					if (p2i != -1) {
						int id = idOf(p1, p2);
						if (!added[id]) {
							added.set(id);
							initialCands.push_back({p1, p2});
						}
					}
				}
			}
		}
		vector<pair<double, array<ivec2, 2>>> candsSorted;
		for (auto [p1, p2] : initialCands) {
			auto& set1 = p.stationToPlaces[p1.pack(N)];
			auto& set2 = p.stationToPlaces[p2.pack(N)];
			double score = 0;
			for (int place1 : set1) {
				for (int place2 : set2) {
					if (place1 == place2)
						continue;
					if ((place1 >> 1) == (place2 >> 1)) {
						score += p.pairDists[place1 >> 1];
					}
				}
			}
			auto addPartial = [&](const vector<int>& places) {
				for (int place1 : places) {
					int place2 = place1 ^ 1;
					ivec2 pos1 = p.poss[place1 >> 1][place1 & 1];
					ivec2 pos2 = p.poss[place2 >> 1][place2 & 1];
					int dist = (pos1 - pos2).mnorm();
					score += dist * initialState.partialIncomeCoeff();
				}
			};
			addPartial(set1);
			addPartial(set2);
			ivec2 diff = (p1 - p2).abs();
			score *= pow(m < 100 ? 0.992 : 0.995, diff.mnorm());
			candsSorted.push_back({score, {p1, p2}});
		}
		ranges::sort(candsSorted, greater<>());
		trace("first pairs determined in ", timer(), " ms. num cands: ", len(candsSorted));

		vector<BsState> bestStates;
		int bestScore = -1;
		vector<int> widths;
		constexpr int MAX_WIDTH = 8000;

		static BsState* buf1 = (BsState*) malloc(sizeof(BsState) * MAX_WIDTH);
		static BsState* buf2 = (BsState*) malloc(sizeof(BsState) * MAX_WIDTH);

		auto runBs = [&](const BsState& initialState, int fixedWidth, int maxStations, bool print) {
			bestStates.clear();
			bestScore = -1;
			widths.clear();
			BsState* src = buf1;
			BsState* dst = buf2;
			int numSrc = 0;
			int numDst = 0;
			static vector<Candidate> cands;
			cands.clear();

			// set initial state
			int numInitialStates = min(500, len(candsSorted));
			rep0(i, numInitialStates) {
				auto [p1, p2] = candsSorted[i].second;
				BsState& st = (src[numSrc++] = initialState);
				// place the first station
				st.updateInsertableCells();
				auto info1 = st.computeNextPlaceInfo(p1, false);
				st.placeStation(info1);
				// place the second station
				st.updateInsertableCells();
				auto info2 = st.computeNextPlaceInfo(p2, false);
				if (info2.score < 0) {
					continue;
				}
				st.placeStation(info2);
			}

			int bsTurn = 2;
			int beamEnd = m < 150 ? 1500 : 1000;
			int baseWidth = (int) round(lerp(2000, 200, linearstep(50, 1600, p.m)));
			int prevWidth = baseWidth;
			beam_width_manager bwm(baseWidth);
			bwm.window_size = 2;

			assert((fixedWidth == -1) != (maxStations == -1));

			if (maxStations == -1)
				maxStations = INF;
			maxStations = min(maxStations, MAX_STATIONS);

			array<ivec2, N2> unpack;
			rep2() {
				unpack[i * N + j] = {i, j};
			}

			while (bsTurn < maxStations) {
				cands.clear();

				auto candDir = [&](const Candidate& a, const Candidate& b) {
					return a.placeInfo.score > b.placeInfo.score;
				};

				// enumerate candidates
				int maxChildrenFromOneParent = 20;
				rep0(index, numSrc) {
					auto& st = src[index];
					int finalScore = st.finalScore();
					if (update_max(bestScore, finalScore)) {
						bestStates.push_back(st);
						if (print && render) {
							mov.target("bs");
							st.draw();
							mov.end_frame();
						}
					}

					static vector<Candidate> localCands;
					localCands.reserve(N2);
					localCands.clear();
					st.updateInsertableCells();
					repn2(posi) {
						if (st.dincomes[posi] > 0) {
							ivec2 pos = unpack[posi];
							auto upperBoundedScore = st.computeNextPlaceInfo(pos, true);
							localCands.push_back({index, upperBoundedScore});
						}
					}
					if (localCands.empty()) {
						repn2(posi) {
							if (st.dpincomes[posi] > 0) {
								ivec2 pos = unpack[posi];
								auto upperBoundedScore = st.computeNextPlaceInfo(pos, true);
								localCands.push_back({index, upperBoundedScore});
							}
						}
					}

					if (len(localCands) > maxChildrenFromOneParent) {
						ranges::nth_element(
							localCands, localCands.begin() + maxChildrenFromOneParent, candDir);
						localCands.resize(maxChildrenFromOneParent);
					}
					cands.insert(cands.end(), localCands.begin(), localCands.end());
				}

				ranges::sort(cands, candDir);

				int numAdded = 0;
				int width;
				if (fixedWidth == -1)
					width = bwm.next(pow(bsTurn / (double) maxStations, 0.5), timer(), beamEnd);
				else
					width = fixedWidth;
				width = min(width, MAX_WIDTH);
				prevWidth = width;
				double lowerBound = -1; // the minimum score to be added to the next states

				auto updateLowerBound = [&]() {
					if (numAdded < width)
						return;
					static vector<double> scores;
					scores.clear();
					rep0(i, numAdded) {
						scores.push_back(cands[i].placeInfo.score);
					}
					ranges::nth_element(scores, scores.begin() + width - 1, greater<>());
					lowerBound = scores[width - 1];
				};

				// filter candiadtes
				static hash_imap<int> bestIndices;
				static hash_imap<int> childrenCount;
				bestIndices.clear();
				childrenCount.clear();
				int numEvaluated = 0;
				int numPruned = 0;
				erase_if(cands, [&](Candidate& cand) {
					if (cand.placeInfo.score <= lowerBound) {
						numPruned++;
						return true;
					}
					int count = 0;
					if (childrenCount.access(cand.parent)) {
						count = childrenCount.get();
						if (count >= maxChildrenFromOneParent) {
							numPruned++;
							return true;
						}
					}
					auto& st = src[cand.parent];
					auto pos = cand.placeInfo.pos;
					auto info = st.computeNextPlaceInfo(pos, false);
					assert(info.score <= cand.placeInfo.score);

					numEvaluated++;
					if (info.score <= lowerBound) {
						numPruned++;
						return true;
					}
					cand.placeInfo = info; // update to the actual score

					// duplicate check
					ull hash = st.nextHash(pos);
					if (bestIndices.access(hash)) {
						// i hope this is safe...
						auto& currentBest = cands[bestIndices.get()];
						// replace if better
						if (currentBest.placeInfo.score < info.score) {
							currentBest.parent = cand.parent;
							currentBest.placeInfo = info;
							updateLowerBound();
						}
						return true; // no need to add this one
					}
					// add to the best indices
					bestIndices.set(numAdded);
					numAdded++;
					updateLowerBound();
					childrenCount.set(count + 1);
					return false;
				});
				if (len(cands) > width) {
					ranges::nth_element(cands, cands.begin() + width, candDir);
					cands.resize(width);
				}
				widths.push_back(len(cands));
				bwm.report(len(cands));
				if (print && bsTurn % 10 == 0) {
					trace("turn=", bsTurn, " cands=", len(cands), " width=", width,
						" evaluated=", numEvaluated, " pruned=", numPruned,
						" lower=", (int) round(lowerBound));
				}

				// check if finished
				if (cands.empty()) {
					break;
				}

				// compute next states
				numDst = 0;
				for (const auto& cand : cands) {
					auto& newNode = (dst[numDst++] = src[cand.parent]);
					newNode.placeStation(cand.placeInfo);
				}
				swap(src, dst);
				swap(numSrc, numDst);
				bsTurn++;
			}
		};

		runBs(initialState, 5, -1, false);
		trace("best score: ", bestScore, " best stations: ", bestStates.back().stations.size());
		int maxStations = (int) (bestStates.back().stations.size() * 1.1 + 20);
		trace("first bs time=", timer());

		runBs(initialState, -1, maxStations, true);

		trace("widths: ", widths);
		trace("best score: ", bestScore, " best stations: ", bestStates.back().stations.size(), "/",
			maxStations);

		// pick the best feasible beam search result
		bool found = [&]() {
			BsState st = initialState;
			while (!bestStates.empty()) {
				int bestScore = st.finalScore();
				rep0(iter, 4) {
					auto st = bestStates.back();
					st.collapse(rng);
					if (render) {
						st.draw();
						mov.comment("best score: " + tos(bestScore));
						mov.end_frame();
					}
					if (!st.computeActions(true).empty()) {
						return true;
					}
				}
				trace("infeasible. picking next best...");
				bestStates.pop_back();
			}
			trace("no feasible solution found");
			return false;
		}();
		assert(msg(found, "WHAT"));

		// extract graph data
		auto st = bestStates.back();
		vector<int> parents(st.stations.size(), -1);
		vector<cvec2> poss(st.stations.size());

		// compute parents by dfs
		parents[0] = -1;
		auto dfs = [&](auto dfs, int index, int parent) -> void {
			parents[index] = parent;
			poss[index] = st.stations[index].pos;
			for (auto& c : st.stations[index].connections) {
				if (c.index != parent) {
					dfs(dfs, c.index, index);
				}
			}
		};
		dfs(dfs, 0, -1);
		trace("-------- bs end\n");

		return {poss, parents};
	}

	void sa(vector<cvec2> poss, vector<int> parents, ostream& out) {
		BsState_old bst; // used for rendering, feasibility checking, and printing
		bst.setup(p, rng);

		SaState st;
		st.setup(p);
		trace("poss: ", poss);
		trace("parents: ", parents);
		st.init(poss, parents);
		st.updateFrom(0);
		auto& nodes = st.nodes;

		int currentScore = st.finalScore();
		int bestScore = currentScore;

		// remove the following lines to enable SA
		// result.score = bestScore;
		// return;

		auto extractData = [&]() {
			poss.clear();
			parents.clear();
			rep0(i, nodes.size()) {
				poss.push_back(nodes[i].pos);
				parents.push_back(nodes[i].parent.to);
			}
		};
		auto hasNoTJunctionFor = [](const vector<int>& parents) {
			static int subIndices[MAX_STATIONS];
			static int smallerCount[MAX_STATIONS];
			int numNodes = len(parents);
			rep0(index, numNodes) {
				subIndices[index] = INF;
				smallerCount[index] = 0;
			}
			subIndices[0] = 0;
			rep(index, 1, numNodes) {
				int pindex = parents[index];
				assert(pindex != -1);
				int baseIndex = index;
				while (subIndices[baseIndex] > index) {
					subIndices[baseIndex] = index;
					baseIndex = parents[baseIndex];
				}
			}
			rep(index, 1, numNodes) {
				int pindex = parents[index];
				if (subIndices[index] < pindex) {
					smallerCount[pindex]++;
					if (smallerCount[pindex] > 1) {
						return false;
					}
				}
			}
			return true;
		};

		auto drawCurrent = [&](string name) {
			mov.target(name);
			extractData();
			bst.sc.placeStations(poss, parents, rng, true);
			bst.draw();
			mov.comment("score: " + tos(currentScore));
			mov.end_frame();

			mov.target(name + "-2");
			st.draw();
			mov.comment("score: " + tos(currentScore));
			mov.end_frame();
		};

		bool verbose = false;

		// best found solutions. may or may not be feasible
		easy_stack<fast_vector<SaState::Node, MAX_STATIONS>> bestNodesList;
		easy_stack<int> bestScoreList;
		bestNodesList.push_back(nodes);
		bestScoreList.push_back(currentScore);

		// the solution to print
		vector<pair<Tile, ivec2>> actionsToPrint;

		// verbose check functions for debugging
		auto verboseCheckScoreInfo = [&](const SaState::ScoreInfo& info, const vector<cvec2>& truePoss,
										 const vector<int>& trueParents) {
			static SaState st3;
			st3.setup(p);
			st3.init(truePoss, trueParents);

			st3.draw();
			mov.comment("correct");
			mov.end_frame();

			auto info2 = st3.updateFrom(0);
			int realScore = info2.finalScore();
			if (info.money != info2.money || info.income != info2.income || info.turn != info2.turn) {
				trace("got: ", info);
				trace("correct: ", info2);
				rep0(from, st3.nodes.size()) {
					auto [money, income, turn] = st.updateFrom(from, true, true);
					trace("from ", from, ": money=", money, " income=", income, " turn=", turn);
				}
				trace("correct data follows:");
				st3.updateFrom(0, false, true);
			}
			assert(info.money == info2.money);
			assert(info.income == info2.income);
			assert(info.turn == info2.turn);
			assert(realScore == info.finalScore());
		};
		auto verboseCheckCurrentScore = [&]() {
			auto oldInfo = st.scores[nodes.size() - 1];
			auto newInfo = st.updateFrom(0, true);
			int recomputedScore = newInfo.finalScore();
			if (recomputedScore != currentScore) {
				trace("currentScore is wrong!!! ", currentScore, " ", recomputedScore);
				trace("oldInfo: ", oldInfo, " score: ", oldInfo.finalScore());
				trace("newInfo: ", newInfo, " score: ", newInfo.finalScore());

				trace("--------------------");

				extractData();
				SaState st3;
				st3.setup(p);
				st3.init(poss, parents);
				auto [money, income, turn] = st3.updateFrom(0);

				rep0(i, nodes.size()) {
					// check node length
					for (auto& c : nodes[i].children) {
						int dist = (nodes[c.to].pos - nodes[i].pos).mnorm();
						if (dist != c.dist) {
							trace("dist mismatch!!! ", dist, " ", c.dist);
						}
					}
					if (i == 0)
						continue;
					int dist = (nodes[nodes[i].parent.to].pos - nodes[i].pos).mnorm();
					if (dist != nodes[i].parent.dist) {
						trace("dist (parent) mismatch!!! ", dist, " ", nodes[i].parent.dist);
					}
				}

				st.updateInfo();
				rep0(from, nodes.size()) {
					auto [money, income, turn] = st.updateFrom(from, true, true);
					trace("from ", from, ": money=", money, " income=", income, " turn=", turn);
				}

				st3.updateFrom(0, true, true);

				trace(DEBUG_STR);

				assert(false);
			}
		};

		// weak feasibility check
		auto canPlaceStations = [&](int center = -1) {
			extractData();
			if (center != -1) {
				static fast_iset<MAX_STATIONS> mask;
				mask.clear();
				mask.insert(center);
				// bfs a bit to include the neighbors
				static queue<int> q;
				while (!q.empty())
					q.pop();
				q.push(center);
				mask.insert(center);
				rep0(iter, 3) {
					int qsize = len(q);
					rep0(j, qsize) {
						int index = q.front();
						q.pop();
						for (auto& c : nodes[index].children) {
							if (mask.insert(c.to)) {
								q.push(c.to);
							}
						}
						if (nodes[index].parent.to != -1) {
							if (mask.insert(nodes[index].parent.to)) {
								q.push(nodes[index].parent.to);
							}
						}
					}
				}
				if (!bst.sc.placeStations(poss, parents, rng, mask))
					return false;
				return true;
			} else {
				if (!bst.sc.placeStations(poss, parents, rng))
					return false;
				return true;
			}
		};

		// util functions for SA transitions
		auto swapNodes = [&](int index1, int index2) {
			if (verbose) {
				mov.target("swap");
				st.draw();
				mov.comment("swapping " + tos(index1) + " " + tos(index2));
				mov.end_frame();
			}

			static fast_iset<MAX_STATIONS + 1> involved;
			involved.clear();
			involved.insert(index1 + 1);
			involved.insert(index2 + 1);
			involved.insert(nodes[index1].parent.to + 1);
			involved.insert(nodes[index2].parent.to + 1);
			for (auto& c : nodes[index1].children) {
				involved.insert(c.to + 1);
			}
			for (auto& c : nodes[index2].children) {
				involved.insert(c.to + 1);
			}
			involved.erase(0);
			for (int i1 : involved) {
				int i = i1 - 1;
				if (st.nodes[i].parent.to == index1) {
					st.nodes[i].parent.to = index2;
				} else if (st.nodes[i].parent.to == index2) {
					st.nodes[i].parent.to = index1;
				}
				for (auto& c : st.nodes[i].children) {
					if (c.to == index1) {
						c.to = index2;
					} else if (c.to == index2) {
						c.to = index1;
					}
				}
			}
			st.uncover(index1);
			st.uncover(index2);
			swap(nodes[index1].pos, nodes[index2].pos);
			swap(nodes[index1].parent, nodes[index2].parent);
			swap(nodes[index1].children, nodes[index2].children);
			swap(nodes[index1].subIndex, nodes[index2].subIndex);
			st.cover(index1);
			st.cover(index2);

			if (verbose) {
				st.draw();
				mov.comment("after swap " + tos(index1) + " " + tos(index2));
				mov.end_frame();
			}

			return min(index1, index2);
		};
		auto changeParent = [&](int index, int newParent) {
			if (verbose) {
				mov.target("changeParent");
				st.draw();
				mov.comment("changing parent of " + tos(index) + " to " + tos(newParent));
				mov.end_frame();
			}

			assert(index > 0);
			assert(nodes[newParent].children.size() + (newParent == 0 ? 0 : 1) < 4);
			int oldParent = nodes[index].parent.to;
			int newDist = (nodes[index].pos - nodes[newParent].pos).mnorm();
			assert(newParent != oldParent);

			nodes[index].parent = {newParent, newDist};
			nodes[oldParent].eraseChild(index);
			nodes[newParent].children.push_back({index, newDist});

			if (verbose) {
				st.draw();
				mov.comment("after changeParent " + tos(index) + " " + tos(newParent));
				mov.end_frame();
			}

			return nodes[index].subIndex;
		};
		auto moveNode = [&](int index, ivec2 newPos) {
			if (verbose) {
				mov.target("moveNode");
				st.draw();
				mov.comment("moving " + tos(index) + " to " + tos(newPos));
				mov.end_frame();
			}

			st.uncover(index);
			nodes[index].pos = newPos;
			st.cover(index);
			if (index > 0) {
				nodes[index].parent.dist = (newPos - nodes[nodes[index].parent.to].pos).mnorm();
				nodes[nodes[index].parent.to].getChild(index).dist = nodes[index].parent.dist;
			}
			int updateFrom = nodes[index].subIndex;
			for (auto& c : nodes[index].children) {
				c.dist = (nodes[c.to].pos - newPos).mnorm();
				nodes[c.to].parent.dist = c.dist;
				updateFrom = min(updateFrom, nodes[c.to].subIndex);
			}
			if (verbose) {
				st.draw();
				mov.comment("after moveNode " + tos(index) + " " + tos(newPos));
				mov.end_frame();
			}
			return updateFrom;
		};

		// individual transitions

		auto trySwapNodes = [&](double tol) {
			int index1;
			int index2;
			while (true) {
				double x;
				x = rng.next_float();
				index1 = (int) ((1 - x * x) * (nodes.size() - 1)) + 1;
				double sd = m < 150 ? 0.2 * len(nodes) : 32;
				index2 = (int) (index1 + rng.next_normal() * sd);
				if (index1 < 1 || index1 >= nodes.size())
					continue;
				if (index2 < 1 || index2 >= nodes.size())
					continue;
				if (index1 == index2)
					continue;
				break;
			}
			int updateFrom = swapNodes(index1, index2);
			auto scoreInfo = st.updateFrom(updateFrom, true);
			int nextScore = scoreInfo.finalScore();
			if (nextScore > currentScore - tol) {
				if (verbose) {
					auto poss2 = poss;
					auto parents2 = parents;
					swap(poss2[index1], poss2[index2]);
					for (int& p : parents2) {
						if (p == index1)
							p = index2;
						else if (p == index2)
							p = index1;
					}
					swap(parents2[index1], parents2[index2]);
					verboseCheckScoreInfo(scoreInfo, poss2, parents2);
				}

				// approve the swap
				int nextScore2 = st.updateFrom(updateFrom).finalScore();
				assert(nextScore == nextScore2);

				if (st.hasNoTJunction()) {
					currentScore = nextScore;
					return true;
				}
				// revert and update
				// trace("sanity check failed! reverting...");
				swapNodes(index1, index2);
				st.updateFrom(updateFrom);
			} else {
				// revert, but no update needed
				swapNodes(index1, index2);
			}
			return false;
		};
		auto tryChangeParent = [&](double tol) {
			int index;
			int currentParent;
			int newParent;
			int trial = 0;
			while (true) {
				if (++trial > 1000)
					return false;
				double x;
				x = rng.next_float();
				index = (int) ((1 - x * x) * (nodes.size() - 1)) + 1;
				if (index == nodes.size())
					continue;
				currentParent = nodes[index].parent.to;
				newParent = rng.next_int(nodes.size());
				if (index == newParent || currentParent == newParent)
					continue;
				int currentDist = (nodes[index].pos - nodes[currentParent].pos).mnorm();
				int newDist = (nodes[index].pos - nodes[newParent].pos).mnorm();
				if ((newDist + 1) >= (currentDist + 1) * 2)
					continue;
				if (nodes[newParent].children.size() + (newParent == 0 ? 0 : 1) == 4)
					continue;
				int heightDiff = nodes[newParent].height - nodes[index].height;
				if (heightDiff > 0) {
					int i = newParent;
					bool looped = false;
					rep0(h, heightDiff) {
						i = nodes[i].parent.to;
						if (i == index) {
							looped = true;
							break;
						}
					}
					if (looped)
						continue;
				}
				break;
			}
			int updateFrom = changeParent(index, newParent);
			auto scoreInfo = st.updateFrom(updateFrom, true);
			int nextScore = scoreInfo.finalScore();
			if (nextScore > currentScore - tol) {
				if (verbose) {
					auto poss2 = poss;
					auto parents2 = parents;
					parents2[index] = newParent;
					verboseCheckScoreInfo(scoreInfo, poss2, parents2);
				}

				// approve the change
				st.updateFrom(updateFrom);

				if (st.hasNoTJunction() && st.hasNoIntersectionForNode(index) &&
					canPlaceStations(newParent)) {
					currentScore = nextScore;
					return true;
				}
				// revert and update
				// trace("sanity check failed! reverting...");
				changeParent(index, currentParent);
				st.updateFrom(updateFrom);
			} else {
				// revert, but no update needed
				changeParent(index, currentParent);
			}
			return false;
		};
		auto tryMoveNode = [&](double tol) {
			int index;
			ivec2 newPos;
			while (true) {
				double x;
				x = rng.next_float();
				index = (int) ((1 - x * x) * nodes.size());
				if (index >= nodes.size())
					continue;
				newPos = nodes[index].pos + ivec2{rng.next_int(-1, 1), rng.next_int(-1, 1)};
				if (newPos == nodes[index].pos || !newPos.in_bounds(N) || st.isOccupied(newPos))
					continue;
				break;
			}
			ivec2 currentPos = nodes[index].pos;
			int updateFrom = moveNode(index, newPos);
			auto scoreInfo = st.updateFrom(updateFrom, true);
			int nextScore = scoreInfo.finalScore();
			if (nextScore > currentScore - tol) {
				if (verbose) {
					auto poss2 = poss;
					auto parents2 = parents;
					poss[index] = newPos;
					verboseCheckScoreInfo(scoreInfo, poss2, parents2);
				}

				// approve the change
				st.updateFrom(updateFrom);

				assert(st.hasNoTJunction());
				if (st.hasNoIntersectionForNode(index, true) && canPlaceStations(index)) {
					currentScore = nextScore;
					return true;
				}
				// revert and update
				moveNode(index, currentPos);
				st.updateFrom(updateFrom);
			} else {
				// revert, but no update needed
				moveNode(index, currentPos);
			}
			return false;
		};
		auto tryInsertNode = [&](double tol) {
			int index;
			int cindex;
			int pindex;
			ivec2 pos;
			while (true) {
				double x;
				x = rng.next_float();
				cindex = (int) ((1 - x * x) * (nodes.size() - 1)) + 1;
				if (cindex >= nodes.size())
					continue;
				index = cindex + (int) (rng.next_normal() * 8);
				if (index <= 0 || index > len(nodes))
					continue;
				pindex = nodes[cindex].parent.to;
				if (rng.next_float() < 0.1) {
					// select completely randomly
					pos = ivec2{rng.next_int(N), rng.next_int(N)};
				} else {
					// near a random point between the parent and the child
					double ratio = rng.next_float();
					ivec2 middle = ivec2{(int) round(lerp(nodes[pindex].pos.i, nodes[cindex].pos.i, ratio)),
						(int) round(lerp(nodes[pindex].pos.j, nodes[cindex].pos.j, ratio))};
					pos = middle + ivec2{(int) (rng.next_normal() * 4), (int) (rng.next_normal() * 4)};
				}
				if (!pos.in_bounds(N) || st.isOccupied(pos))
					continue;
				break;
			}
			auto info = st.computeScoreForNodeInsertion(index, cindex, pos);
			int nextScore = info.finalScore();
			if (nextScore > currentScore - tol) {
				static vector<array<ivec2, 2>> segments;
				segments.clear();
				segments.push_back({nodes[cindex].pos, pos});
				segments.push_back({pos, nodes[pindex].pos});
				if (!st.hasNoIntersectionForSegments(segments)) {
					// impossible...
					return false;
				}
				// validate the move
				extractData();

				auto oldPoss = poss;
				auto oldParents = parents;

				poss.insert(poss.begin() + index, pos);
				parents.insert(parents.begin() + index, pindex);
				for (int& p : parents) {
					if (p >= index)
						p++;
				}
				parents[cindex + (cindex >= index)] = index;
				if (!bst.sc.placeStations(poss, parents, rng) || !hasNoTJunctionFor(parents)) {
					// impossible...
					return false;
				}

				st.init(poss, parents);
				st.updateFrom(0);

				assert(st.hasNoTJunction());
				assert(st.hasNoIntersectionForNode(index, true));
				assert(st.finalScore() == nextScore);
				currentScore = nextScore;
				return true;
			}
			return false;
		};
		auto tryConnectNode = [&](double tol) {
			int pindex;
			int index;
			ivec2 pos;
			while (true) {
				double x;
				x = rng.next_float();
				pindex = (int) ((1 - x * x) * nodes.size()) + 1;
				if (pindex >= nodes.size())
					continue;
				index = pindex + (int) (rng.next_normal() * 8);
				if (index <= 0 || index > len(nodes))
					continue;
				if (nodes[pindex].numConnections() >= 4)
					continue;
				if (nodes[pindex].subIndex < nodes[pindex].index && index <= pindex)
					continue; // forbid obvious T-junctions; still possible to happen so check later as well
				if (rng.next_float() < 0.1) {
					// select completely randomly
					pos = ivec2{rng.next_int(N), rng.next_int(N)};
				} else {
					// near the parent
					pos = nodes[pindex].pos +
						ivec2{(int) (rng.next_normal() * 4), (int) (rng.next_normal() * 4)};
				}
				if (!pos.in_bounds(N) || st.isOccupied(pos))
					continue;
				break;
			}
			auto info = st.computeScoreForNodeConnection(index, pindex, pos);
			int nextScore = info.finalScore();
			if (nextScore > currentScore - tol) {
				static vector<array<ivec2, 2>> segments;
				segments.clear();
				segments.push_back({pos, nodes[pindex].pos});
				if (!st.hasNoIntersectionForSegments(segments)) {
					// impossible...
					return false;
				}
				// validate the move
				extractData();
				poss.insert(poss.begin() + index, pos);
				parents.insert(parents.begin() + index, pindex);
				parents[index] = pindex;
				for (int& p : parents) {
					if (p >= index)
						p++;
				}
				if (!bst.sc.placeStations(poss, parents, rng) || !hasNoTJunctionFor(parents)) {
					// impossible...
					return false;
				}

				st.init(poss, parents);
				st.updateFrom(0);

				assert(st.hasNoTJunction());
				assert(st.hasNoIntersectionForNode(index, true));
				assert(st.finalScore() == nextScore);
				assert(canPlaceStations(index));
				currentScore = nextScore;
				return true;
			}
			return false;
		};
		auto tryRemoveNode = [&](double tol) {
			int index;
			int pindex;
			while (true) {
				index = rng.next_int(1, nodes.size() - 1);
				pindex = nodes[index].parent.to;
				if (nodes[pindex].numConnections() + nodes[index].children.size() - 1 > 4)
					continue;
				int smallerCount = 0;
				for (auto& c : nodes[index].children) {
					if (nodes[c.to].subIndex < nodes[pindex].index) {
						smallerCount++;
					}
				}
				if (smallerCount > 1)
					continue; // forbid T-junctions
				break;
			}

			ivec2 pos = nodes[index].pos;
			ivec2 ppos = nodes[pindex].pos;
			st.indexToIgnore = index;
			st.uncover(index);
			for (auto& c : nodes[index].children) {
				nodes[c.to].parent = {pindex, (nodes[c.to].pos - ppos).mnorm()};
			}
			auto info = st.updateFrom(nodes[index].subIndex, true);
			st.indexToIgnore = -1;
			st.cover(index);
			for (auto& c : nodes[index].children) {
				nodes[c.to].parent = {index, (nodes[c.to].pos - pos).mnorm()};
			}
			int nextScore = info.finalScore();
			if (nextScore > currentScore - tol) {
				static vector<array<ivec2, 2>> segments;
				segments.clear();
				for (auto& c : nodes[index].children) {
					segments.push_back({nodes[c.to].pos, ppos});
				}
				if (!st.hasNoIntersectionForSegments(segments)) {
					// impossible...
					return false;
				}

				// validate the move
				extractData();
				poss.erase(poss.begin() + index);
				parents.erase(parents.begin() + index);
				for (int& p : parents) {
					if (p == index)
						p = pindex;
					if (p > index)
						p--;
				}
				if (!bst.sc.placeStations(poss, parents, rng)) {
					// impossible...
					return false;
				}

				st.init(poss, parents);
				st.updateFrom(0);

				assert(st.hasNoTJunction());
				assert(st.hasNoIntersectionForNode(pindex - (pindex > index), true));
				assert(st.finalScore() == nextScore);
				assert(canPlaceStations());
				currentScore = nextScore;
				return true;
			}
			return false;
		};
		auto tryNormalize = [&]() {
			static vector<cvec2> tmpPoss;
			static vector<int> tmpParents;
			extractData();
			tmpPoss = poss;
			tmpParents = parents;
			if (st.normalize()) {
				if (!canPlaceStations()) {
					st.init(tmpPoss, tmpParents);
					st.updateFrom(0);
					assert(st.finalScore() == currentScore);
					return false;
				}
				currentScore = st.scores[nodes.size() - 1].finalScore();
				return true;
			}
			return false;
		};
		auto switchRoot = [&]() {
			// switch the root: 0 <-> 1
			// give chance for 0 to change
			poss.resize(len(nodes));
			parents.resize(len(nodes));
			auto dfs = [&](auto dfs, int index, int parent) -> void {
				parents[index] = parent;
				poss[index] = nodes[index].pos;
				for (auto& c : nodes[index].children) {
					if (c.to != parent) {
						dfs(dfs, c.to, index);
					}
				}
				if (index > 0) {
					if (nodes[index].parent.to != parent) {
						dfs(dfs, nodes[index].parent.to, index);
					}
				}
			};
			// make 1 the new root
			dfs(dfs, 1, -1);
			// swap 0 and 1
			swap(poss[0], poss[1]);
			swap(parents[0], parents[1]);
			rep0(i, len(nodes)) {
				if (parents[i] == 0)
					parents[i] = 1;
				else if (parents[i] == 1)
					parents[i] = 0;
			}
			st.init(poss, parents);
			st.updateFrom(0);
			assert(st.hasNoTJunction());
			assert(st.hasNoIntersectionForNode(0, true) && st.hasNoIntersectionForNode(1, true));
			assert(st.finalScore() == currentScore);
		};

		double progress = 0;
		double temp;
		int noHit = 0;
		int iter = 0;
		int start = min(2800, timer());
		double tempFrom = lerp(1000, 5000, linearstep(50, 1600, p.m));
		double tempTo = 50;
		int restoreInterval = (int) round(lerp(50000, 10000, linearstep(50, 1600, p.m)));
		Selector sel;

		struct TranStat {
			int accept = 0;
			int total = 0;
		};
		array<TranStat, 7> stats;

		auto tryTransition = [&](double tol) {
			if (iter % 1024 == 0) {
				// no need to check the score
				switchRoot();
			}
			switch (sel.select(rng)) {
			case 0:
				return len(nodes) > 3 && trySwapNodes(tol);
			case 1:
				return len(nodes) > 3 && tryChangeParent(tol);
			case 2:
				return tryMoveNode(tol);
			case 3:
				return len(nodes) < MAX_STATIONS && tryInsertNode(tol);
			case 4:
				return len(nodes) < MAX_STATIONS && tryConnectNode(tol);
			case 5:
				return len(nodes) > 3 && tryRemoveNode(tol);
			case 6:
				return tryNormalize();
			default:
				assert(false);
			}
		};

		// checks if the best is updated
		int progressSteps = 0;
		double lastBestPrinted = 0;
		int lastBestPrintedTime = 0;

		auto checkBestUpdate = [&]() {
			if (currentScore <= bestScore)
				return; // nope

			// try to always normalize
			tryNormalize();
			// must not contain intersections
			// assert(st.hasNoSelfIntersections());
			// and not have any forbidden structures
			assert(st.hasNoTJunction());

			noHit = 0; // hit!
			bestScore = currentScore;
			bestNodesList.push_back(nodes);
			bestScoreList.push_back(bestScore);

			if (bestScore > lastBestPrinted + 1000 || timer() - lastBestPrintedTime > 100) {
				lastBestPrinted = bestScore;
				lastBestPrintedTime = timer();
				if (progress > 0.5)
					trace("new best score: ", bestScore, " iter=", iter, " temp=", temp);
			}

			// ensure feasbiility
			if (progress * 10 > progressSteps + 1) {
				trace("checking feasibility...");
				progressSteps = (int) (progress * 10);

				// is current best feasible?
				while (!bestNodesList.empty()) {
					extractData();
					bool feasible = false;
					rep0(iter, 4) {
						if (!bst.sc.placeStations(poss, parents, rng)) {
							continue;
						}
						bst.sc.collapse(rng);
						if (!bst.sc.computeActions(true).empty()) {
							feasible = true;
							break;
						}
					}
					if (feasible) {
						break; // okay!
					}

					if (render) {
						mov.target("discard");
						int stuckAt = 0;
						rep(i, 1, len(nodes)) {
							bool ok = false;
							for (auto& c : bst.sc.stations[i].connections) {
								if (c.index == parents[i]) {
									ok = true;
									break;
								}
							}
							if (!ok) {
								stuckAt = i;
								break;
							}
						}
						bst.draw();
						if (stuckAt > 0) {
							mov.fill(1, 0, 1, 0.4);
							mov.no_stroke();
							mov.circle(poss[stuckAt].j + 0.5, poss[stuckAt].i + 0.5, 2);
							mov.comment("stuck at " + tos(stuckAt) + " (pos=" + tos(poss[stuckAt]) + ")");
						}
						mov.end_frame();
						SaState st2;
						st2.setup(p);
						st2.init(poss, parents);
						st2.draw();
						mov.end_frame();
					}

					// discard
					bestNodesList.pop_back();
					bestScoreList.pop_back();
					trace("discarded best");
					// revert to the previous best
					nodes = bestNodesList.back();
					currentScore = bestScore = bestScoreList.back();
					st.recomputeDeltaIncome();
					st.updateFrom(0);
					st.updateInfo();
				}
				assert(msg(!bestNodesList.empty(), "the first one must be feasible"));
				trace("ok!");
			}
			if (render) {
				drawCurrent("sa-best");
			}
		};

		trace("SA started with ", currentScore, " time=", start);
		for (; iter < 100000000; iter++) {
			if ((iter & 0xff) == 0) {
				progress = linearstep(start, 2950, timer());
				if (progress >= 1)
					break;
				temp = exp_interp(tempFrom, tempTo, progress);
				// update the selector
				sel.clear();
				auto interp = [&](array<double, 2> params) {
					return lerp(params[0], params[1], progress);
				};
				sel.add(interp(params.trans0));
				sel.add(interp(params.trans1));
				sel.add(interp(params.trans2));
				sel.add(interp(params.trans3));
				sel.add(interp(params.trans4));
				sel.add(interp(params.trans5));
				sel.add(interp(params.trans6));
			}
			if ((iter & 0xffff) == 0) {
				trace("current score: ", currentScore, " iter=", iter, " temp=", temp);
				if (render) {
					drawCurrent("sa-current");
				}
			}

			// restore the best state if no progress is made for a while
			if (++noHit > restoreInterval) {
				nodes = bestNodesList.back();
				st.recomputeDeltaIncome();
				st.updateFrom(0);
				st.updateInfo();
				currentScore = bestScore;
				noHit = 0;
				trace("restored best iter=", iter, " temp=", temp);
			}

			double tol = -log(rng.next_float()) * temp;

			if (verbose) {
				extractData();
			}
			// try transtion. update the best if succeeded
			if (tryTransition(tol)) {
				stats[sel.last].accept++;
				checkBestUpdate();
				st.updateInfo();
			}
			stats[sel.last].total++;
		}

		// print stats
		trace("transition stats:");
		string transNames[] = {
			"swap", "changeParent", "moveNode", "insertNode", "connectNode", "removeNode", "normalize"};
		rep0(i, 7) {
			trace(transNames[i], ": ", stats[i].accept, " / ", stats[i].total, " (",
				100.0 * stats[i].accept / stats[i].total, "%)");
		}

		auto printFinalSolution = [&]() {
			if (render)
				mov.target("final");
			while (!bestNodesList.empty()) {
				auto bestNodes = bestNodesList.pop_back();
				int bestScore = bestScoreList.pop_back();

				poss.clear();
				parents.clear();
				rep0(i, bestNodes.size()) {
					poss.push_back(bestNodes[i].pos);
					parents.push_back(bestNodes[i].parent.to);
				}

				rep0(iter, 4) {
					if (!bst.sc.placeStations(poss, parents, rng, true)) {
						continue;
					}
					bst.sc.collapse(rng);

					if (render) {
						nodes = bestNodes;
						st.recomputeDeltaIncome();
						st.updateFrom(0);
						st.normalize();
						mov.comment("connection");
						st.draw();
						mov.end_frame();
						bst.draw();
						mov.comment("after collapse");
						mov.end_frame();
					}
					auto actions = bst.sc.computeActions();
					if (!actions.empty()) {
						printSolution(out, actions);
						trace("solution printed! stations=", len(poss));
						result.score = bestScore;
						return true;
					}
					if (render) {
						bst.draw();
						mov.comment("failed");
						mov.end_frame();
					}
				}
				trace("infeasible. picking next best...");
			}
			trace("no feasible SA solution found");
			return false; // :(
		};

		bool forceFinalRender = false;

		bool prevRender = render;
		render = render || forceFinalRender;
		if (!printFinalSolution()) {
			assert(msg(false, "WHAT"));
		}
		render = prevRender;
	}

	void printSolution(ostream& out, const vector<pair<Tile, ivec2>>& actions) {
		assert(!actions.empty());
		int money = p.k;
		int income = 0;
		int turn = 0;
		bool placeCovered[MAX_M * 2] = {};
		bool pairFulfilled[MAX_M] = {};
		int numPairsFulfilled = 0;
		auto waitForMoney = [&](int target) {
			assert(money >= target || income > 0);
			while (money < target) {
				money += income;
				turn++;
				out << "-1" << endl;
			}
		};
		for (auto [t, pos] : actions) {
			if (t == Tile::station()) {
				waitForMoney(STATION_COST);
				money -= STATION_COST;
				// fulfill pairs
				for (int place : p.stationToPlaces[pos.pack(N)]) {
					// trace("covered place: ", place);
					if (!placeCovered[place]) {
						placeCovered[place] = true;
						if (placeCovered[place ^ 1]) {
							int pair = place >> 1;
							pairFulfilled[pair] = true;
							income += p.pairDists[pair];
							numPairsFulfilled++;
							// trace("pair fulfilled: ", pair);
						}
					}
				}
				money += income;
				turn++;
				out << 0 << " " << pos.i << " " << pos.j << endl;
				// trace("built a station at ", pos, ". money: ", money, " income: ", income);
			} else {
				waitForMoney(RAIL_COST);
				money -= RAIL_COST;
				money += income;
				turn++;
				out << t.kind() << " " << pos.i << " " << pos.j << endl;
				// trace("built a rail at ", pos, ". money: ", money, " income: ", income);
			}
		}
		while (turn < T) {
			turn++;
			out << "-1" << endl;
		}
	}
};

int main(int argc, char* argv[]) {
#if 0 || ONLINE_JUDGE
	isLocal = false;
	render = false;
	timer(true);
	Solver sol;
	sol.load(cin);
	sol.solve();
#elif 0
	// for masters
	makeMovie(argc, argv);
#elif 0
	// write metadata
	ostringstream oss;
	oss << "seed m k" << endl;
	rep0(seed, 5000) {
		Solver sol;
		sol.load(seed);
		double mlog = log2(sol.p.m / 50.0);
		double mLogApprox = (int) (round(mlog * 10)) / 10.0;
		int mApprox = (int) (round(pow(2, mLogApprox) * 50));
		int kApprox = (int) (round(sol.p.k / 500.0) * 500);
		oss << seed << " " << mApprox << " " << kApprox << endl;
	}
	write_text("scores/input.txt", oss.str());
#elif 1
	// for local/remote testers
	debug = false;
	render = false;
	int seed;
	cin >> seed;
	cin >> time_scale;

	Params params;

	// cin >> params.trans0[0] >> params.trans0[1];
	// cin >> params.trans1[0] >> params.trans1[1];
	// cin >> params.trans2[0] >> params.trans2[1];
	// cin >> params.trans3[0] >> params.trans3[1];
	// cin >> params.trans4[0] >> params.trans4[1];
	// cin >> params.trans5[0] >> params.trans5[1];
	// cin >> params.trans6[0] >> params.trans6[1];

	timer(true);
	Solver sol;
	sol.params = params;
	sol.load(seed);
	sol.solve();

	cout << sol.result.score << " " << timer() << endl;
#elif 1
	// single-threaded test, handy but slow
	int num = 10;
	int from = 0;
	int stride = 1;
	int single = -1;

	vector<int> seedList = {};

	debug = true;
	render = false;

	if (render)
		time_scale = 2.0;

	struct TestCase {
		int seed;
		int time;
		ll score;
	};
	vector<TestCase> cases;

	if (seedList.empty() || single != -1) {
		seedList.clear();
		if (single == -1) {
			rep0(t, num) {
				seedList.push_back(from + t * stride);
			}
		} else {
			seedList.push_back(single);
		}
	}

	bool doTrace = debug;
	debug = true;
	for (int seed : seedList) {
		timer(true);
		trace("------------ SOLVING SEED ", seed, " ------------");
		debug = doTrace;
		Solver s;
		s.load(seed);
		s.solve();
		debug = true;
		int time = timer();
		trace("score: ", s.result.score, " (time ", time, " ms)\n");
		if (s.result.score != -1)
			cases.emplace_back(seed, time, s.result.score);
	}

	auto print = [&](const TestCase& c) {
		int seed = c.seed;
		string space = seed < 10 ? "   " : seed < 100 ? "  " : seed < 1000 ? " " : "";
		trace("  seed ", space, seed, ": ", c.score, " (time ", c.time, " ms)");
	};

	if (len(cases) > 1) {
		trace("------------ summary ------------");

		trace("sort by score:");
		sort(cases.begin(), cases.end(), [&](auto a, auto b) {
			return a.score > b.score;
		});
		for (auto& c : cases)
			print(c);

		trace("sort by seed:");
		sort(cases.begin(), cases.end(), [&](auto a, auto b) {
			return a.seed < b.seed;
		});
		for (auto& c : cases)
			print(c);

		ll scoreSum = 0;
		double logScoreSum = 0;
		for (auto& c : cases) {
			scoreSum += c.score;
			logScoreSum += log(c.score);
		}
		double invDenom = 1.0 / len(cases);
		trace("total score: ", scoreSum, ", mean: ", (ll) (scoreSum * invDenom * 100 + 0.5) / 100.0,
			", mean(log2): ", (ll) (logScoreSum * invDenom * 1000 + 0.5) / 1000.0);
	}
#endif
}
