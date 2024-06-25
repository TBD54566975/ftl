include!(concat!(env!("OUT_DIR"), "/call_immediate.rs"));

fn main() {
    ftl::builder::main(call_immediate)
}
