


rule Rule {
	meta:
 		severity = "HIGH"
   	strings:
      	$trigger1 = "Invalid_Rule_1-1"
		$trigger2 = "Invalid_Rule_1-2"
   	condition:
      	1 of them
}


rule Rule {
    meta:
        severity = "HIGH"
    strings:
        $trigger = "Valid_Rule_4"
    condition:
        $trigger
}