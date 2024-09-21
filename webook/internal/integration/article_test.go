package integration

import (
	"github.com/stretchr/testify/suite"
	"testing"
)

// ArticleTestSuite 测试套件
type ArticleTestSuite struct {
	suite.Suite
}

func (s *ArticleTestSuite) TestABC() {
	s.T().Log("HELLO")
}

func TestArticle(t *testing.T) {
	suite.Run(t, &ArticleTestSuite{})
}
