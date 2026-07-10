rule Valid_Rule_1 {
    meta:
        severity = "HIGH"
    strings:
        $trigger = "Valid_Rule_1"
    condition:
        $trigger
}

rule Invalid_Rule_1 {
	meta:
 		severity = "HIGH"
   	strings:
      	$trigger1 = "Invalid_Rule_1-1"
		$trigger2 = "Invalid_Rule_1-2"
   	condition:
      	1 of them
}

rule Valid_Rule_2 {
    meta:
        severity = "HIGH"
    strings:
        $trigger = "Valid_Rule_2"
    condition:
        $trigger
}
rule Valid_Rule_3 {
	meta:
		severity = "HIGH"
	strings:
		$trigger = "Valid_Rule_3"
	condition:
		$trigger
}
