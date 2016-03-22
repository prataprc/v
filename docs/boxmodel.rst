invariants followed in box model
--------------------------------

* border, padding, content shall be rendered from outside to inside.

* border is specified as:
    <type>,<fgcolor:fgattribute>,<fgcolor:fgattribute>
* type can be "line", or "none"
* color can be one of the eight color or integer value less than 256
* attribute can be "bold", "underline", "reverse"

* padding is specifed in three different ways
  a. <top>,<right>,<bottom>,<left>
  b. <top>,<right> where <bottom>,<left> is same as <top>,<right>
  c. <top> where <right>,<bottom>,<left> is same as <top>

* content is screen left within the padded box.
* every box will be rooted at (x,y) coordinate with reference to terminal.
* once aligned width and height of the box shall be fixed.
* if width is specified and exceeds the containing box's width, it will be
  adjusted, to fit the containing box.
* if height is specified and exceeds the containing box's width, it will be
  adjusted, to fit the containing box.
