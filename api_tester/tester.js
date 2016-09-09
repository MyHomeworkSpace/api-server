$(document).ready(function() {
	$("#addParameter").click(function() {
		var $param = $('<li class="param"></li>');
			var $removeBtn = $('<button>-</button>');
				$removeBtn.click(function() {
					$(this).parent().remove();
				});
			$param.append($removeBtn);
			var $keyText = $('<input type="text" class="keyText" value="Key">');
			$param.append($keyText);
			var $valueText = $('<input type="text" class="valueText" value="Value">');
			$param.append($valueText);
		$("#paramContainer").append($param);
	});
	$("#submit").click(function() {
		var method = $("#method").val();
		var path = $("#path").val();
		var parameters = {};
		$("#paramContainer li").each(function() {
			parameters[$(this).children(".keyText").val()] = $(this).children(".valueText").val();
		});
		var csrfToken = "";
		var cookies = document.cookie.split(";");
		for (var cookieIndex in cookies) {
			if (cookies[cookieIndex].trim().startsWith("csrfToken=")) {
				csrfToken = cookies[cookieIndex].trim().replace("csrfToken=", "");
			}
		}
		if (csrfToken === "") {
			alert("Failed to get csrfToken. Does it exist?");
			return;
		}
		var finalPath = path + "?csrfToken=" + encodeURIComponent(csrfToken);
		$("#responseLine").text("Loading...");
		$("#response").text("");
		$.ajax({
			data: parameters,
			method: method,
			url: finalPath,
			complete: function(xhr, status) {
				$("#responseLine").text(xhr.status + " " + xhr.statusText);
				try {
					$("#response").text(JSON.stringify(JSON.parse(xhr.responseText), null, 4));
				} catch (e) {
					// malformed JSON?
					$("#response").text(xhr.responseText);
				}
			}
		})
	});
});
