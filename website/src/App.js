import React, { useState, useEffect } from "react";
import CssBaseline from "@mui/material/CssBaseline";
import Grid from "@mui/material/Grid";
import Container from "@mui/material/Container";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import Header from "./Components/Header";
import MainFeaturedPost from "./Components/MainFeaturedPost";
import FeaturedPost from "./Components/FeaturedPost";
import Footer from "./Components/Footer";

const sections = [
  { title: "Cloud", url: "#" },
  { title: "DevOps", url: "#" },
];

// TODO remove, this demo shouldn't need to reset the theme.
const App = () => {
  const [mainFeaturedPost, setMainFeaturedPost] = useState([]);
  const [featuredPosts, setFeaturedPosts] = useState([]);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const response = await fetch(`${process.env.REACT_APP_API_URL}/posts`);
        const data = await response.json();
        setMainFeaturedPost(data.shift());
        setFeaturedPosts(data);
      } catch (error) {
        console.error("Error fetching data:", error);
      }
    };

    fetchData();
  }, []);

  return (
    <ThemeProvider theme={createTheme()}>
      <CssBaseline />
      <Container maxWidth='lg'>
        <Header title='Crawler System Testing 1' sections={sections} />
        <main>
          {mainFeaturedPost.map((post, idx) => (
            <MainFeaturedPost key={idx} post={post} />
          ))}
          <Grid container spacing={4}>
            {featuredPosts.map((post, idx) => (
              <FeaturedPost key={post.title} index={idx} post={post} />
            ))}
          </Grid>
        </main>
      </Container>
      <Footer title='Footer' description='Demo Crawler System' />
    </ThemeProvider>
  );
};

export default App;
