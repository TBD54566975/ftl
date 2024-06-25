include!(concat!(env!("OUT_DIR"), "/call_immediate.rs"));

fn main() {
    ftl::runner::main(call_immediate)
}
