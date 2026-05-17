use shopify_function::prelude::*;

#[typegen("./schema.graphql")]
pub mod schema {
    #[query("./src/run.graphql")]
    pub mod cart_delivery_options_transform_run {}
}

pub mod run;
