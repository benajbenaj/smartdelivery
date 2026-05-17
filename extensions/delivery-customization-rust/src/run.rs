use super::schema;
use shopify_function::prelude::*;
use shopify_function::Result;

#[shopify_function]
fn cart_delivery_options_transform_run(
    _input: schema::cart_delivery_options_transform_run::Input,
) -> Result<schema::CartDeliveryOptionsTransformRunResult> {
    Ok(schema::CartDeliveryOptionsTransformRunResult {
        operations: Vec::new(),
    })
}
