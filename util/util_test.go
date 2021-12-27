package util

import "testing"

func TestArticleText(t *testing.T) {
	title := "LocationMacedonia.png"
	text := `14:46 2004年4月30日 [[利用者:Oxhop|Oxhop]] &#34;[[:画像:LocationMacedonia.png|LocationMacedonia.png]]&#34;をアップロードしました。 (マケドニアの位置 - 英語版より)`

	actual := ParseArticle(title, text)

	expected := `14:46 2004年4月30日 Oxhop "LocationMacedonia.png"をアップロードしました。 (マケドニアの位置 - 英語版より)`

	if actual != expected {
		t.Fatalf("expected: %s, actual: %s", expected, actual)
	}
}
