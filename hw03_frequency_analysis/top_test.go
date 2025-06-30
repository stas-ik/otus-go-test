package hw03frequencyanalysis

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// Change to true if needed.
var taskWithAsteriskIsCompleted = false

var text = `Как видите, он  спускается  по  лестнице  вслед  за  своим
	другом   Кристофером   Робином,   головой   вниз,  пересчитывая
	ступеньки собственным затылком:  бум-бум-бум.  Другого  способа
	сходить  с  лестницы  он  пока  не  знает.  Иногда ему, правда,
		кажется, что можно бы найти какой-то другой способ, если бы  он
	только   мог   на  минутку  перестать  бумкать  и  как  следует
	сосредоточиться. Но увы - сосредоточиться-то ему и некогда.
		Как бы то ни было, вот он уже спустился  и  готов  с  вами
	познакомиться.
	- Винни-Пух. Очень приятно!
		Вас,  вероятно,  удивляет, почему его так странно зовут, а
	если вы знаете английский, то вы удивитесь еще больше.
		Это необыкновенное имя подарил ему Кристофер  Робин.  Надо
	вам  сказать,  что  когда-то Кристофер Робин был знаком с одним
	лебедем на пруду, которого он звал Пухом. Для лебедя  это  было
	очень   подходящее  имя,  потому  что  если  ты  зовешь  лебедя
	громко: "Пу-ух! Пу-ух!"- а он  не  откликается,  то  ты  всегда
	можешь  сделать вид, что ты просто понарошку стрелял; а если ты
	звал его тихо, то все подумают, что ты  просто  подул  себе  на
	нос.  Лебедь  потом  куда-то делся, а имя осталось, и Кристофер
	Робин решил отдать его своему медвежонку, чтобы оно не  пропало
	зря.
		А  Винни - так звали самую лучшую, самую добрую медведицу
	в  зоологическом  саду,  которую  очень-очень  любил  Кристофер
	Робин.  А  она  очень-очень  любила  его. Ее ли назвали Винни в
	честь Пуха, или Пуха назвали в ее честь - теперь уже никто  не
	знает,  даже папа Кристофера Робина. Когда-то он знал, а теперь
	забыл.
		Словом, теперь мишку зовут Винни-Пух, и вы знаете почему.
		Иногда Винни-Пух любит вечерком во что-нибудь поиграть,  а
	иногда,  особенно  когда  папа  дома,  он больше любит тихонько
	посидеть у огня и послушать какую-нибудь интересную сказку.
		В этот вечер...`

func TestTop10(t *testing.T) {
	t.Run("no words in empty string", func(t *testing.T) {
		require.Len(t, Top10(""), 0)
	})

	t.Run("positive test", func(t *testing.T) {
		if taskWithAsteriskIsCompleted {
			expected := []string{
				"а",         // 8
				"он",        // 8
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"в",         // 4
				"его",       // 4
				"если",      // 4
				"кристофер", // 4
				"не",        // 4
			}
			require.Equal(t, expected, Top10(text))
		} else {
			expected := []string{
				"он",        // 8
				"а",         // 6
				"и",         // 6
				"ты",        // 5
				"что",       // 5
				"-",         // 4
				"Кристофер", // 4
				"если",      // 4
				"не",        // 4
				"то",        // 4
			}
			require.Equal(t, expected, Top10(text))
		}
	})

	t.Run("punctuation as part of words", func(t *testing.T) {
		input := "cat, cat dog, dog! dog dog."
		expected := []string{
			"cat",  // 1
			"cat,", // 1
			"dog",  // 1
			"dog!", // 1
			"dog,", // 1
			"dog.", // 1
		}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("dash as separate word", func(t *testing.T) {
		input := "word1 - word2 - word3 word1 - word1"
		expected := []string{
			"-",     // 3
			"word1", // 3
			"word2", // 1
			"word3", // 1
		}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("case sensitivity", func(t *testing.T) {
		input := "Cat cat CAT cAt"
		expected := []string{
			"CAT", // 1
			"Cat", // 1
			"cAt", // 1
			"cat", // 1
		}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("equal frequency lexicographic sort", func(t *testing.T) {
		input := "zebra apple banana cat dog eagle"
		expected := []string{
			"apple",  // 1
			"banana", // 1
			"cat",    // 1
			"dog",    // 1
			"eagle",  // 1
			"zebra",  // 1
		}
		require.Equal(t, expected, Top10(input))
	})

	t.Run("more than 10 words with same frequency", func(t *testing.T) {
		input := "a b c d e f g h i j k l m"
		expected := []string{
			"a", // 1
			"b", // 1
			"c", // 1
			"d", // 1
			"e", // 1
			"f", // 1
			"g", // 1
			"h", // 1
			"i", // 1
			"j", // 1
		}
		require.Equal(t, expected, Top10(input))
	})

}
