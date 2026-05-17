use super::schema;
use serde::Deserialize;
use shopify_function::prelude::*;
use shopify_function::Result;

#[derive(Debug, Default, Deserialize, PartialEq)]
pub struct FunctionConfig {
    #[serde(default)]
    rules: Vec<ConfigRule>,
}

#[derive(Debug, Deserialize, PartialEq)]
struct ConfigRule {
    condition: ConfigPredicate,
    action: ConfigPredicate,
}

#[derive(Debug, Deserialize, PartialEq)]
struct ConfigPredicate {
    r#type: String,
    value: String,
}

struct CartContext {
    subtotal_amount: String,
    delivery_groups: Vec<DeliveryGroupContext>,
}

struct DeliveryGroupContext {
    country_code: Option<String>,
    zip: Option<String>,
    delivery_options: Vec<DeliveryOptionContext>,
}

struct DeliveryOptionContext {
    handle: String,
    title: Option<String>,
    code: Option<String>,
    cost_amount: String,
}

#[shopify_function]
fn cart_delivery_options_transform_run(
    input: schema::cart_delivery_options_transform_run::Input,
) -> Result<schema::CartDeliveryOptionsTransformRunResult> {
    let config: &FunctionConfig = match input.delivery_customization().metafield() {
        Some(metafield) => metafield.json_value(),
        None => return Ok(no_changes()),
    };

    let cart = cart_context(&input);
    let operations = hidden_delivery_option_handles(config, &cart)
        .into_iter()
        .map(|handle| {
            schema::Operation::DeliveryOptionHide(schema::DeliveryOptionHideOperation {
                delivery_option_handle: handle,
            })
        })
        .collect();

    Ok(schema::CartDeliveryOptionsTransformRunResult { operations })
}

fn no_changes() -> schema::CartDeliveryOptionsTransformRunResult {
    schema::CartDeliveryOptionsTransformRunResult {
        operations: Vec::new(),
    }
}

fn cart_context(input: &schema::cart_delivery_options_transform_run::Input) -> CartContext {
    let cart = input.cart();
    let subtotal_amount = cart.cost().subtotal_amount().amount().to_string();
    let delivery_groups = cart
        .delivery_groups()
        .iter()
        .map(|group| {
            let delivery_options = group
                .delivery_options()
                .iter()
                .map(|option| DeliveryOptionContext {
                    handle: option.handle().to_string(),
                    title: option.title().as_ref().map(|title| title.to_string()),
                    code: option.code().as_ref().map(|code| code.to_string()),
                    cost_amount: option.cost().amount().to_string(),
                })
                .collect();

            DeliveryGroupContext {
                country_code: group
                    .delivery_address()
                    .and_then(|address| {
                        address
                            .country_code()
                            .as_ref()
                            .map(|code| code.to_string())
                    }),
                zip: group
                    .delivery_address()
                    .and_then(|address| address.zip().as_ref().map(|zip| zip.to_string())),
                delivery_options,
            }
        })
        .collect();

    CartContext {
        subtotal_amount,
        delivery_groups,
    }
}

fn hidden_delivery_option_handles(config: &FunctionConfig, cart: &CartContext) -> Vec<String> {
    let mut handles = Vec::new();

    for rule in &config.rules {
        if rule.action.r#type != "hide" {
            continue;
        }

        for group in &cart.delivery_groups {
            if !condition_matches(&rule.condition, cart.subtotal_amount.as_str(), group) {
                continue;
            }

            for option in &group.delivery_options {
                if delivery_option_matches(&rule.action.value, option)
                    && !handles.iter().any(|handle| handle == &option.handle)
                {
                    handles.push(option.handle.clone());
                }
            }
        }
    }

    handles
}

fn condition_matches(
    condition: &ConfigPredicate,
    cart_subtotal_amount: &str,
    group: &DeliveryGroupContext,
) -> bool {
    match condition.r#type.as_str() {
        "country" => {
            optional_eq_ignore_ascii_case(group.country_code.as_deref(), &condition.value)
        }
        "postal_code" => postal_code_matches(group.zip.as_deref(), &condition.value),
        "cart_total" => amount_matches(cart_subtotal_amount, &condition.value),
        _ => false,
    }
}

fn delivery_option_matches(target: &str, option: &DeliveryOptionContext) -> bool {
    let target = target.trim();
    if target.is_empty() {
        return false;
    }

    if let Some(value) = target.strip_prefix("title:") {
        return optional_contains_ignore_ascii_case(option.title.as_deref(), value.trim());
    }
    if let Some(value) = target.strip_prefix("code:") {
        return optional_eq_ignore_ascii_case(option.code.as_deref(), value.trim());
    }
    if let Some(value) = target.strip_prefix("cost:") {
        return amount_matches(option.cost_amount.as_str(), value.trim());
    }

    optional_contains_ignore_ascii_case(Some(option.handle.as_str()), target)
        || optional_contains_ignore_ascii_case(option.title.as_deref(), target)
        || optional_contains_ignore_ascii_case(option.code.as_deref(), target)
        || target.eq_ignore_ascii_case("free") && amount_matches(option.cost_amount.as_str(), "0")
}

fn optional_eq_ignore_ascii_case(value: Option<&str>, expected: &str) -> bool {
    value
        .map(|value| value.eq_ignore_ascii_case(expected.trim()))
        .unwrap_or(false)
}

fn optional_contains_ignore_ascii_case(value: Option<&str>, expected: &str) -> bool {
    let expected = expected.trim().to_lowercase();
    value
        .map(|value| value.to_lowercase().contains(&expected))
        .unwrap_or(false)
}

fn postal_code_matches(actual: Option<&str>, expected: &str) -> bool {
    let Some(actual) = actual.map(normalize_postal_code) else {
        return false;
    };

    expected
        .split(',')
        .map(str::trim)
        .filter(|part| !part.is_empty())
        .any(|part| postal_code_part_matches(&actual, part))
}

fn postal_code_part_matches(actual: &str, expected: &str) -> bool {
    if let Some((start, end)) = expected.split_once('-') {
        let start = normalize_postal_code(start);
        let end = normalize_postal_code(end);
        return actual >= start.as_str() && actual <= end.as_str();
    }

    actual == normalize_postal_code(expected)
}

fn normalize_postal_code(value: &str) -> String {
    value
        .chars()
        .filter(|char| !char.is_whitespace())
        .flat_map(char::to_uppercase)
        .collect()
}

fn amount_matches(actual: &str, expected: &str) -> bool {
    let Some(actual) = parse_amount(actual) else {
        return false;
    };
    let expected = expected.trim();

    if let Some(value) = expected.strip_prefix(">=") {
        return parse_amount(value).map(|value| actual >= value).unwrap_or(false);
    }
    if let Some(value) = expected.strip_prefix("<=") {
        return parse_amount(value).map(|value| actual <= value).unwrap_or(false);
    }
    if let Some(value) = expected.strip_prefix('>') {
        return parse_amount(value).map(|value| actual > value).unwrap_or(false);
    }
    if let Some(value) = expected.strip_prefix('<') {
        return parse_amount(value).map(|value| actual < value).unwrap_or(false);
    }
    if let Some((start, end)) = expected.split_once('-') {
        return match (parse_amount(start), parse_amount(end)) {
            (Some(start), Some(end)) => actual >= start && actual <= end,
            _ => false,
        };
    }

    parse_amount(expected)
        .map(|expected| (actual - expected).abs() < f64::EPSILON)
        .unwrap_or(false)
}

fn parse_amount(value: &str) -> Option<f64> {
    value.trim().parse::<f64>().ok()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn hides_option_when_postal_code_and_title_match() {
        let config = config("postal_code", "10000-19999", "hide", "pickup");
        let cart = cart("120.00", Some("US"), Some("10012"), vec![option(
            "pickup-1",
            Some("Local Pickup"),
            Some("PICKUP"),
            "0.00",
        )]);

        let handles = hidden_delivery_option_handles(&config, &cart);

        assert_eq!(handles, vec!["pickup-1"]);
    }

    #[test]
    fn does_not_hide_when_address_condition_fails() {
        let config = config("country", "US", "hide", "express");
        let cart = cart("120.00", Some("CA"), Some("H2X 1Y4"), vec![option(
            "express-1",
            Some("Express Shipping"),
            Some("EXPRESS"),
            "20.00",
        )]);

        let handles = hidden_delivery_option_handles(&config, &cart);

        assert!(handles.is_empty());
    }

    #[test]
    fn matches_option_by_code_and_cost() {
        let config = FunctionConfig {
            rules: vec![
                rule("country", "US", "hide", "code:EXPRESS"),
                rule("country", "US", "hide", "cost:0"),
            ],
        };
        let cart = cart(
            "120.00",
            Some("US"),
            Some("10012"),
            vec![
                option("standard-1", Some("Standard"), Some("STANDARD"), "5.00"),
                option("express-1", Some("Express"), Some("EXPRESS"), "20.00"),
                option("pickup-1", Some("Pickup"), Some("PICKUP"), "0.00"),
            ],
        );

        let handles = hidden_delivery_option_handles(&config, &cart);

        assert_eq!(handles, vec!["express-1", "pickup-1"]);
    }

    #[test]
    fn supports_cart_total_condition() {
        let config = config("cart_total", ">=100", "hide", "free");
        let cart = cart("120.00", Some("US"), Some("10012"), vec![option(
            "free-1",
            Some("Free Shipping"),
            Some("FREE"),
            "0.00",
        )]);

        let handles = hidden_delivery_option_handles(&config, &cart);

        assert_eq!(handles, vec!["free-1"]);
    }

    fn config(
        condition_type: &str,
        condition_value: &str,
        action_type: &str,
        action_value: &str,
    ) -> FunctionConfig {
        FunctionConfig {
            rules: vec![rule(
                condition_type,
                condition_value,
                action_type,
                action_value,
            )],
        }
    }

    fn rule(
        condition_type: &str,
        condition_value: &str,
        action_type: &str,
        action_value: &str,
    ) -> ConfigRule {
        ConfigRule {
            condition: ConfigPredicate {
                r#type: condition_type.to_string(),
                value: condition_value.to_string(),
            },
            action: ConfigPredicate {
                r#type: action_type.to_string(),
                value: action_value.to_string(),
            },
        }
    }

    fn cart(
        subtotal_amount: &str,
        country_code: Option<&str>,
        zip: Option<&str>,
        delivery_options: Vec<DeliveryOptionContext>,
    ) -> CartContext {
        CartContext {
            subtotal_amount: subtotal_amount.to_string(),
            delivery_groups: vec![DeliveryGroupContext {
                country_code: country_code.map(str::to_string),
                zip: zip.map(str::to_string),
                delivery_options,
            }],
        }
    }

    fn option(
        handle: &str,
        title: Option<&str>,
        code: Option<&str>,
        cost_amount: &str,
    ) -> DeliveryOptionContext {
        DeliveryOptionContext {
            handle: handle.to_string(),
            title: title.map(str::to_string),
            code: code.map(str::to_string),
            cost_amount: cost_amount.to_string(),
        }
    }
}
