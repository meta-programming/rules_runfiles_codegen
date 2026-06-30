# Simple templating helper for Starlark with validation.

def render(template, **kwargs):
    """Replaces {{key}} placeholders in the template with their string values.

    This avoids the need for complex escaping of curly braces (which are common
    in Go/Kotlin) or percent signs (common in Go format verbs) that standard
    Starlark formatting operators require.

    Performs strict validation:
    - Fails if a key in kwargs is not found in the template (unused argument).
    - Fails if any {{placeholder}} remains in the template after rendering (missing argument).
    """
    result = template

    # 1. Check for unused arguments and perform replacement
    for k, v in kwargs.items():
        placeholder = "{{" + k + "}}"
        if result.find(placeholder) == -1:
            fail("Template does not contain placeholder: %s" % placeholder)
        result = result.replace(placeholder, str(v))

    # 2. Check for unresolved placeholders
    if result.find("{{") != -1:
        start = result.find("{{")
        end = result.find("}}", start)
        if end != -1:
            placeholder = result[start:end + 2]
            fail("Unresolved placeholder in template: %s (did you forget to pass it?)" % placeholder)
        else:
            fail("Malformed placeholder (missing closing '}}') in template starting at: %s" % result[start:start + 20])

    return result
